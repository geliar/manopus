package processor

import (
	"context"
	"sync"

	"github.com/geliar/manopus/pkg/payload"
)

type catalogStore struct {
	processors map[string]Processor
	sync.RWMutex
}

var catalog catalogStore

func Register(ctx context.Context, processor Processor) {
	catalog.register(ctx, processor)
}

func Run(ctx context.Context, config *ProcessorConfig, payload *payload.Payload) (result interface{}, err error) {
	return catalog.run(ctx, config, payload)
}

func (c *catalogStore) register(ctx context.Context, processor Processor) {
	c.Lock()
	defer c.Unlock()
	l := logger(ctx)
	if c.processors == nil {
		l.Debug().
			Msg("Initializing processor catalog")
		c.processors = make(map[string]Processor)
	}
	l = l.With().
		Str("processor_type", processor.Type()).
		Logger()
	if _, ok := c.processors[processor.Type()]; ok {
		l.Fatal().
			Msg("Cannot register processor with existing name")
	}
	c.processors[processor.Type()] = processor
	l.Info().
		Msg("Registered new processor")
}

func (c *catalogStore) run(ctx context.Context, config *ProcessorConfig, payload *payload.Payload) (result interface{}, err error) {
	c.RLock()

	l := logger(ctx)
	if _, ok := c.processors[config.Type]; !ok {
		c.RUnlock()
		l.Error().
			Str("processor_type", config.Type).
			Msgf("Cannot find processor with type '%s'", config.Type)
		return
	}
	p := c.processors[config.Type]
	c.RUnlock()
	return p.Run(ctx, config, payload)
}
