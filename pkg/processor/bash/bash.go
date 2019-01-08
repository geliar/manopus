package bash

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/geliar/manopus/pkg/log"

	"github.com/geliar/manopus/pkg/payload"
	"github.com/geliar/manopus/pkg/processor"
)

func init() {
	ctx := log.Logger.WithContext(context.Background())
	l := logger(ctx)
	l.Debug().Msg("Registering processor in the catalog")
	processor.Register(ctx, new(Bash))
}

type Bash struct {
}

func (p *Bash) Type() string {
	return serviceName
}

func (p *Bash) Run(ctx context.Context, config *processor.ProcessorConfig, payload *payload.Payload) (result interface{}, next processor.NextStatus, err error) {
	next = processor.NextContinue
	l := logger(ctx)
	script := p.collectScript(ctx, config.Script)
	if script == "" {
		return nil, processor.NextStopSequence, processor.ErrParseScript
	}

	cmd := exec.CommandContext(ctx, "/bin/bash", "/dev/stdin")
	cmd.Env = os.Environ()
	stdin, err := cmd.StdinPipe()
	if err != nil {
		l.Error().Err(err).Msg("Cannot open stdin of executing process")
		return nil, processor.NextStopSequence, err
	}
	stdout, err := cmd.StdoutPipe()
	var stdoutBuf []byte
	stdoutReadCh := make(chan struct{})
	go func() {
		stdoutBuf, err = ioutil.ReadAll(stdout)
		if err != nil {
			l.Error().Err(err).Msg("Cannot read stdout of executing process")
		}
		stdoutReadCh <- struct{}{}
	}()
	if err != nil {
		l.Error().Err(err).Msg("Cannot open stdout of executing process")
		return nil, processor.NextStopSequence, err
	}
	go func() {
		<-ctx.Done()
		_ = stdin.Close()
		_ = stdout.Close()
	}()
	pp := preparePayload("ENV_", payload.Env)
	pp += preparePayload("MATCH_", payload.Match)
	pp += preparePayload("EXPORT_", payload.Export)
	pp += preparePayload("REQ_", payload.Req)
	go func() {
		_, _ = stdin.Write([]byte(pp))
		_, _ = stdin.Write([]byte(script))
		_ = stdin.Close()
	}()
	err = cmd.Start()
	if err != nil {
		l.Error().Err(err).Msg("Cannot start the script")
		return nil, processor.NextStopSequence, err
	}

	err = cmd.Wait()
	if err != nil {
		next = processor.NextStopSequence
		l.Debug().Err(err).Msg("Error when executing script")
	}
	<-stdoutReadCh
	return string(stdoutBuf), next, err
}

func (Bash) collectScript(ctx context.Context, script interface{}) (result string) {
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

func preparePayload(prefix string, payload map[string]interface{}) (result string) {
	var builder strings.Builder
	for k, p := range payload {
		key := prefix + k
		switch v := p.(type) {
		case string, int, float64:
			builder.WriteString(prepareKey(key))
			builder.WriteString(`="`)
			builder.WriteString(prepareValue(fmt.Sprint(v)))
			builder.WriteString("\"\n")
		case []interface{}:
			for i := range v {
				switch a := v[i].(type) {
				case string, int, float64:
					builder.WriteString(prepareKey(key))
					builder.WriteString(fmt.Sprintf(`[%d]="`, i))
					builder.WriteString(prepareValue(fmt.Sprint(a)))
					builder.WriteString("\"\n")
				}
			}
		case map[string]interface{}:
			builder.WriteString(preparePayload(key+"_", v))
		}
	}
	return builder.String()
}

func prepareKey(str string) string {
	str = strings.Replace(str, " ", "_", -1)
	str = strings.Replace(str, "-", "_", -1)
	str = strings.ToUpper(str)
	return str
}

func prepareValue(str string) string {
	str = strings.Replace(str, `"`, `\"`, -1)
	return str
}
