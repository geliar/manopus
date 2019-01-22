package python

import (
	"context"
	"encoding/json"
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
	processor.Register(log.Logger.WithContext(context.Background()), new(Python))
}

type Python struct {
}

func (p *Python) Type() string {
	return serviceName
}

func (p *Python) Run(ctx context.Context, config *processor.ProcessorConfig, payload *payload.Payload) (result interface{}, next processor.NextStatus, err error) {
	next = processor.NextContinue
	l := logger(ctx)
	script := p.collectScript(ctx, config.Script)
	if script == "" {
		return nil, processor.NextStopSequence, processor.ErrParseScript
	}
	script = `import json
f = open("/dev/stdin")
payload = json.load(f)
` + script
	tmp, err := ioutil.TempFile("", serviceName)
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err != nil {
		l.Error().Err(err).Msg("Cannot create temporary script file")
		return nil, processor.NextStopSequence, err
	}
	_, err = tmp.Write([]byte(script))
	if err != nil {
		l.Error().Err(err).Msg("Cannot write to temporary script file")
		return nil, processor.NextStopSequence, err
	}

	cmd := exec.CommandContext(ctx, "python", tmp.Name())
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
	payloadBuf, err := json.Marshal(payload)
	go func() {
		_, _ = stdin.Write([]byte(payloadBuf))
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
		l.Error().Err(err).Msg("Error when executing script")
	}
	<-stdoutReadCh
	return string(stdoutBuf), next, err
}

func (Python) collectScript(ctx context.Context, script interface{}) (result string) {
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
