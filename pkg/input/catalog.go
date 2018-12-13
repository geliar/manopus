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

func Register(name string, driver Driver) {
	catalog.register(name, driver)
}

func StopAll(ctx context.Context) {
	catalog.stopAll(ctx)
}

func (c *catalogStore) register(name string, driver Driver) {
	c.Lock()
	defer c.Unlock()
	l := logger()
	if c.inputs == nil {
		l.Debug().
			Msg("Initializing input catalog")
		c.inputs = make(map[string]Driver)
	}
	l = l.With().Str("input_driver_name", name).Logger()
	if _, ok := c.inputs[name]; ok {
		l.Fatal().
			Msg("Trying to register input driver with existing name")
	}
	c.inputs[name] = driver
	l.Info().
		Msg("Registered new input driver")
}

func (c *catalogStore) stopAll(ctx context.Context) {
	c.Lock()
	defer c.Unlock()
	for i := range c.inputs {
		c.inputs[i].Stop(ctx)
	}
}
