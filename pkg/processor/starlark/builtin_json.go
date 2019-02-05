package starlark

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DLag/starlight/convert"
	"go.starlark.net/starlark"
)

func NewJson(ctx context.Context) starlark.Value {
	return NewBuiltIn(ctx, map[string]builtinGetter{
		"parse": jsonParse,
		"dump":  jsonDump,
	})
}

func jsonParse(ctx context.Context) *starlark.Builtin {
	fname := "json.parse"
	l := logger(ctx).With().Str("starlark_function", fname).Logger()
	return starlark.NewBuiltin("rand.int", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if args.Len() != 1 || args.Index(0).Type() != "string" {
			l.Error().Int("args_len", args.Len()).Msgf("Wrong args. Should be %s(string).", fname)
			return starlark.None, fmt.Errorf("wrong args, should be %s(string)", fname)
		}
		buf := args.Index(0).(starlark.String).GoString()
		var v map[string]interface{}
		err := json.Unmarshal([]byte(buf), &v)
		if err != nil {
			l := logger(ctx)
			l.Error().
				Err(err).
				Msg("Error when unmarshaling json")
			return starlark.None, err
		}
		return convert.MakeDict(v)
	})
}

func jsonDump(ctx context.Context) *starlark.Builtin {
	fname := "json.dump"
	l := logger(ctx).With().Str("starlark_function", fname).Logger()
	return starlark.NewBuiltin("rand.int", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if args.Len() != 1 || args.Index(0).Type() != "dict" {
			l.Error().Int("args_len", args.Len()).Msgf("Wrong args. Should be %s(dict).", fname)
			return starlark.None, fmt.Errorf("wrong args, should be %s(dict)", fname)
		}
		d := convertToStringMap(convert.FromDict(args.Index(0).(*starlark.Dict)))
		buf, err := json.Marshal(d)
		if err != nil {
			l := logger(ctx)
			l.Error().
				Err(err).
				Msg("Error when marshaling json")
			return starlark.None, err
		}
		return toValue(string(buf))
	})
}
