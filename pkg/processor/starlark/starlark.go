package starlark

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	sljson "github.com/DLag/starlark-modules/json"
	slrandom "github.com/DLag/starlark-modules/random"
	"github.com/DLag/starlight/convert"
	"github.com/pkg/errors"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"

	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/output"
	"github.com/geliar/manopus/pkg/payload"
	"github.com/geliar/manopus/pkg/processor"
	"github.com/geliar/manopus/pkg/report"
)

func init() {
	processor.Register(log.Logger.WithContext(context.Background()), new(Starlark))
	rand.Seed(time.Now().UTC().UnixNano())
	resolve.AllowGlobalReassign = true
}

type Starlark struct {
}

func (p *Starlark) Type() string {
	return serviceName
}

func (p *Starlark) Run(ctx context.Context, reporter report.Driver, rawScript interface{}, event *payload.Event, payload *payload.Payload) (next processor.NextStatus, callback interface{}, responses []payload.Response, err error) {
	next = processor.NextContinue
	script := p.collectScript(ctx, rawScript)
	var result *bool
	result, callback, responses, err = p.run(ctx, reporter, script, event, payload)
	if result == nil && err == nil {
		next = processor.NextContinue
		return
	}
	if *result {
		next = processor.NextRepeatStep
		return
	}
	next = processor.NextStopSequence
	return
}

//Match execution of match
func (p Starlark) Match(ctx context.Context, rawMatch interface{}, payload *payload.Payload) (matched bool, err error) {
	script := p.collectScript(ctx, rawMatch)
	matched, err = p.match(ctx, script, payload)
	return
}

func (p Starlark) run(ctx context.Context, reporter report.Driver, script string, event *payload.Event, pl *payload.Payload) (result *bool, respond interface{}, responses []payload.Response, err error) {
	l := logger(ctx)
	l.Debug().Str("script", script).
		Msg("Executing script")
	globals := p.makeGlobals(ctx, pl)
	globals["report"] = func(v string) {
		log.Debug().
			Str("starlark_function", "report").
			Str("param-v", v).
			Msg("Called function")
		reporter.PushString(ctx, v)
	}
	globals["respond"] = func(response interface{}) {
		l.Debug().
			Str("starlark_function", "respond").
			Msg("Received response from script")
		if respond != nil {
			l.Warn().
				Str("starlark_function", "respond").
				Msg("Got several respond calls from script, using data from last one")
		}
		respond = response
	}
	globals["send"] = starlark.NewBuiltin("send", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		l := l.With().Str("starlark_function", "send").Logger()
		if args.Len() != 2 || args.Index(0).Type() != "string" || args.Index(1).Type() != "dict" {
			l.Error().Int("args_len", args.Len()).Msg("Wrong args. Should be send(string, dict).")
			return starlark.None, errors.New("wrong args, should be send(string, dict).")
		}
		outputName := args.Index(0).(starlark.String).GoString()
		response := convert.FromDict(args.Index(1).(*starlark.Dict))
		l.Debug().
			Str("output_name", outputName).
			Msg("Received response from script")
		converted := convertToStringMap(response).(map[string]interface{})
		responses = append(responses, payload.Response{Output: outputName, Data: converted})
		return starlark.None, nil
	})
	globals["call"] = starlark.NewBuiltin("call", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		l := l.With().Str("starlark_function", "call").Logger()
		if args.Len() != 2 || args.Index(0).Type() != "string" || args.Index(1).Type() != "dict" {
			l.Error().Int("args_len", args.Len()).Msg("Wrong args. Should be call(string, dict).")
			return starlark.None, errors.New("wrong args, should be call(string, dict).")
		}
		outputName := args.Index(0).(starlark.String).GoString()
		response := convert.FromDict(args.Index(1).(*starlark.Dict))
		l.Debug().
			Str("output_name", outputName).
			Msg("Received call request from script")
		converted := convertToStringMap(response).(map[string]interface{})
		res := output.Send(ctx, &payload.Response{ID: event.ID, Output: outputName, Data: converted, Request: event})
		if res == nil {
			return starlark.None, nil
		}
		v, err := toValue(res)
		if err != nil {
			l.Error().Err(err).Msg("Cannot convert output response to Starlark value")
			return starlark.None, err
		}
		return v, nil
	})
	globals["repeat"] = func() {
		l.Debug().
			Str("starlark_function", "repeat").
			Msg("Script asked to repeat the sequence")
		r := true
		result = &r
	}
	globals["stop"] = func() {
		l.Debug().
			Str("starlark_function", "stop").
			Msg("Script asked to stop the sequence")
		r := false
		result = &r
	}
	dict, err := convert.MakeStringDict(globals)
	if err != nil {
		l.Error().Err(err).Msg("Error converting payload to Starlark globals")
		r := false
		return &r, nil, nil, err
	}
	th := &starlark.Thread{}
	_, err = starlark.ExecFile(th, "manopus_script.star", script, dict)
	if err != nil {
		l.Error().Err(err).Msg("Error executing Starlark script")
		r := false
		return &r, nil, nil, err
	}
	pl.Export = convertToStringMap(convert.FromDict(globals["export"].(*starlark.Dict))).(map[string]interface{})
	if result != nil {
		l.Debug().Msgf("Script execution result is %t", *result)
	}
	return result, respond, responses, nil
}

func (p Starlark) match(ctx context.Context, script string, pl *payload.Payload) (matched bool, err error) {
	l := logger(ctx)
	l.Debug().
		Msgf("Matching with Starlark")
	globals := p.makeGlobals(ctx, pl)
	globals["matched"] = func(b bool) {
		l.Debug().
			Str("starlark_function", "matched").
			Msgf("Match script called matched with %t", b)
		matched = b
	}
	dict, err := convert.MakeStringDict(globals)
	if err != nil {
		l.Error().
			Err(err).
			Msg("Error converting payload to Starlark globals")
		return false, nil
	}
	thread := &starlark.Thread{}
	sf, pr, err := starlark.SourceProgram("manopus_match.star", script, dict.Has)
	if err != nil {
		l.Error().Err(err).Msg("Error parsing Starlark match")
		return false, err
	}
	exp := p.soleExpr(sf)
	if exp != nil {
		v, err := starlark.EvalExpr(thread, exp, dict)
		if err != nil {
			l.Debug().Err(err).Msg("Error executing Starlark expression")
			return false, err
		}
		if b, ok := v.(starlark.Bool); ok {
			matched = bool(b)
		}
		return matched, nil
	}
	_, err = pr.Init(thread, dict)

	if err != nil {
		l.Error().Err(err).Msg("Error executing Starlark match")
		return false, err
	}

	return
}

func (Starlark) collectScript(ctx context.Context, script interface{}) (result string) {
	l := logger(ctx)
	switch v := script.(type) {
	case []interface{}:
		var builder strings.Builder
		for i := range v {
			switch s := v[i].(type) {
			case string, int, float64:
				builder.WriteString(fmt.Sprint(s))
				builder.WriteString("\n")
			default:
				l.Error().Msgf("Cannot parse script in line %d, skipping", i)
				return
			}
		}
		return builder.String()
	case string:
		return v + "\n"
	}
	return ""
}

func (Starlark) soleExpr(f *syntax.File) syntax.Expr {
	if len(f.Stmts) == 1 {
		if stmt, ok := f.Stmts[0].(*syntax.ExprStmt); ok {
			return stmt.X
		}
	}
	return nil
}

func (p Starlark) makeGlobals(ctx context.Context, payload *payload.Payload) map[string]interface{} {
	l := logger(ctx)
	env, _ := toValue(payload.Env)
	vars, _ := toValue(payload.Vars)
	req, _ := toValue(payload.Req)
	export, _ := toValue(payload.Export)
	match, _ := toValue(payload.Match)
	return map[string]interface{}{
		"env":    env,
		"vars":   vars,
		"req":    req,
		"export": export,
		"match":  match,
		"sleep": func(duration int) {
			c := time.After(time.Duration(duration) * time.Millisecond)
			select {
			case <-ctx.Done():
				return
			case <-c:
				return
			}
		},
		"json": sljson.New(),
		"match_re": starlark.NewBuiltin("match_re", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			//func(str string, re string) (matched bool) {
			l := l.With().Str("starlark_function", "match_re").Logger()
			if args.Len() != 2 || args.Index(0).Type() != "string" || args.Index(1).Type() != "string" {
				l.Error().Int("args_len", args.Len()).Msg("Wrong args. Should be match_re(string, string).")
				return starlark.None, errors.New("wrong args, should be match_re(string, string)")
			}
			str := args.Index(0).(starlark.String).GoString()
			re := args.Index(1).(starlark.String).GoString()
			l.Debug().
				Str("param-str", str).
				Str("param-re", re).
				Msg("Called function")
			r, err := regexp.Compile(re)
			if err != nil {
				l := logger(ctx)
				l.Error().
					Err(err).
					Msg("Error when compiling regexp")
				return starlark.False, err
			}
			if !r.MatchString(str) {
				return starlark.False, nil
			}

			if payload.Match == nil {
				payload.Match = make(map[string]interface{})
			}
			results := r.FindStringSubmatch(str)
			names := r.SubexpNames()
			for i, m := range results {
				if i != 0 && names[i] != "" {
					payload.Match[names[i]] = m
					sk, err := toValue(names[i])
					if err != nil {
						continue
					}
					sv, err := toValue(m)
					if err != nil {
						continue
					}
					_ = match.(*starlark.Dict).SetKey(sk, sv)
				}
			}
			return starlark.True, nil
		}),
		"var_get": starlark.NewBuiltin("var_get", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			l := l.With().Str("starlark_function", "var_get").Logger()
			if args.Len() != 1 || args.Index(0).Type() != "string" {
				l.Error().Int("args_len", args.Len()).Msg("Wrong args. Should be var_get(string).")
				return starlark.None, errors.New("wrong args, should be var_get(string)")
			}
			query := args.Index(0).(starlark.String).GoString()
			l.Debug().
				Str("param-query", query).
				Msg("Called function")
			v, err := toValue(payload.QueryField(ctx, query))
			if err != nil {
				l.Error().Msg("Error converting received value")
				return starlark.None, errors.New("error converting received value")
			}

			l.Debug().Str("query", query).
				Str("starlark_function", "var_get").
				Msg("Getting value from payload")
			return v, nil
		}),
		"debug": func(v interface{}) {
			log.Debug().
				Str("starlark_function", "debug").
				Str("param-v", fmt.Sprint(v)).
				Msg("Called function")
		},
		"random": slrandom.New(),
	}
}
