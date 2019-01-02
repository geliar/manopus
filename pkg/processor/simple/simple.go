package simple

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/BurntSushi/toml"

	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/payload"
	"github.com/geliar/manopus/pkg/processor"
)

func init() {
	ctx := log.Logger.WithContext(context.Background())
	l := logger(ctx)
	l.Debug().Msg("Registering processor in the catalog")
	processor.Register(ctx, new(Simple))
}

type Simple struct {
}

func (p *Simple) Type() string {
	return serviceName
}

func (p *Simple) Run(ctx context.Context, config *processor.ProcessorConfig, payload *payload.Payload) (result interface{}, next processor.NextStatus, err error) {
	next = processor.NextContinue
	l := logger(ctx)
	switch config.Encoding {
	case "none":
		return config.Script, next, err
	case "json":
		buf, err := json.Marshal(config.Script)
		if err != nil {
			l.Error().Err(err).Msg("Error marshaling script field to JSON")
			return nil, processor.NextStopSequence, err
		}
		return buf, next, err
	case "toml":
		var buf bytes.Buffer
		enc := toml.NewEncoder(bufio.NewWriter(&buf))
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = errors.New("TOML Encoder panicked")
				}
			}()
			err = enc.Encode(config.Script)
		}()
		if err != nil {
			l.Error().Err(err).Msg("Error marshaling script field to TOML")
			return nil, processor.NextStopSequence, err
		}
		return buf.Bytes(), next, err
	}
	l.Error().
		Str("processor_encoding", config.Encoding).
		Msg("Unsupported encoding")
	return nil, processor.NextStopSequence, errors.New("unsupported encoding")
}
