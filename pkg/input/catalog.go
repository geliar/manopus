package input

import (
	"context"
	"sync"
)

type catalogStore struct {
	inputs map[string]Driver
	sync.RWMutex
}

var catalog catalogStore

func Register(ctx context.Context, name string, driver Driver) {
	catalog.register(ctx, name, driver)
}

func RegisterHandlerAll(ctx context.Context, handler Handler) {
	catalog.registerHandlerAll(ctx, handler)
}

func StopAll(ctx context.Context) {
	catalog.stopAll(ctx)
}

func (c *catalogStore) register(ctx context.Context, name string, driver Driver) {
	c.Lock()
	defer c.Unlock()
	l := logger(ctx)
	if c.inputs == nil {
		l.Debug().
			Msg("Initializing input catalog")
		c.inputs = make(map[string]Driver)
	}
	l = l.With().
		Str("input_driver_name", name).
		Str("input_driver_type", driver.Type()).
		Logger()
	if _, ok := c.inputs[name]; ok {
		l.Fatal().
			Msg("Cannot register input driver with existing name")
	}
	c.inputs[name] = driver
	l.Info().
		Msg("Registered new input driver")
}

func (c *catalogStore) registerHandlerAll(ctx context.Context, handler Handler) {
	c.RLock()
	defer c.RUnlock()
	l := logger(ctx)

	for i := range c.inputs {
		c.inputs[i].RegisterHandler(ctx, handler)
		l.Debug().
			Str("input_driver_name", c.inputs[i].Name()).
			Str("input_driver_type", c.inputs[i].Type()).
			Msgf("Registered handler to input")
	}
}

func (c *catalogStore) stopAll(ctx context.Context) {
	c.Lock()
	defer c.Unlock()
	l := logger(ctx)

	for i := range c.inputs {
		l.Info().
			Str("input_driver_name", c.inputs[i].Name()).
			Str("input_driver_type", c.inputs[i].Type()).
			Msgf("Shutting down input driver")
		c.inputs[i].Stop(ctx)
		delete(c.inputs, i)
	}
}
