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

func Run(ctx context.Context, name string, script interface{}, payload *payload.Payload) (next NextStatus, callback interface{}, responses []payload.Response, err error) {
	return catalog.run(ctx, name, script, payload)
}

func Match(ctx context.Context, name string, match interface{}, payload *payload.Payload) (matched bool, err error) {
	return catalog.match(ctx, name, match, payload)
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
		Str("processor_name", processor.Type()).
		Logger()
	if _, ok := c.processors[processor.Type()]; ok {
		l.Fatal().
			Msg("Cannot register processor with existing name")
	}
	c.processors[processor.Type()] = processor
	l.Debug().
		Msg("Registered new processor")
}

func (c *catalogStore) run(ctx context.Context, name string, script interface{}, payload *payload.Payload) (next NextStatus, callback interface{}, responses []payload.Response, err error) {
	c.RLock()

	l := logger(ctx)
	if _, ok := c.processors[name]; !ok {
		c.RUnlock()
		l.Error().
			Str("processor_name", name).
			Msgf("Cannot find processor with name '%s'", name)
		return
	}
	p := c.processors[name]
	c.RUnlock()
	return p.Run(ctx, script, payload)
}

func (c *catalogStore) match(ctx context.Context, name string, match interface{}, payload *payload.Payload) (matched bool, err error) {
	c.RLock()

	l := logger(ctx)
	if _, ok := c.processors[name]; !ok {
		c.RUnlock()
		l.Error().
			Str("processor_name", name).
			Msgf("Cannot find processor with name '%s'", name)
		return
	}
	p := c.processors[name]
	c.RUnlock()
	return p.Match(ctx, match, payload)
}
