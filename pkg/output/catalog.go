package output

import (
	"context"
	"sync"
)

type catalogStore struct {
	outputs map[string]Driver
	sync.RWMutex
}

var catalog catalogStore

func Register(ctx context.Context, name string, driver Driver) {
	catalog.register(ctx, name, driver)
}

func Send(ctx context.Context, output string, response *Response) {
	catalog.send(ctx, output, response)
}

func StopAll(ctx context.Context) {
	catalog.stopAll(ctx)
}

func (c *catalogStore) register(ctx context.Context, name string, driver Driver) {
	c.Lock()
	defer c.Unlock()
	l := logger(ctx)
	if c.outputs == nil {
		l.Debug().
			Msg("Initializing output catalog")
		c.outputs = make(map[string]Driver)
	}
	l = l.With().
		Str("output_driver_name", name).
		Str("output_driver_type", driver.Type()).
		Logger()
	if _, ok := c.outputs[name]; ok {
		l.Fatal().
			Msg("Cannot register output driver with existing name")
	}
	c.outputs[name] = driver
	l.Info().
		Msg("Registered new output driver")
}

func (c *catalogStore) send(ctx context.Context, output string, response *Response) {
	c.RLock()
	l := logger(ctx)
	if _, ok := c.outputs[output]; !ok {
		c.RUnlock()
		l.Error().
			Str("output_driver_name", output).
			Msgf("Cannot find output driver with name '%s'", output)
		return
	}
	o := c.outputs[output]
	c.RUnlock()
	o.Send(ctx, response)
}

func (c *catalogStore) stopAll(ctx context.Context) {
	c.Lock()
	defer c.Unlock()
	l := logger(ctx)

	for i := range c.outputs {
		l.Info().
			Str("output_driver_name", c.outputs[i].Name()).
			Str("output_driver_type", c.outputs[i].Type()).
			Msgf("Stopping output driver")
		c.outputs[i].Stop(ctx)
		delete(c.outputs, i)
	}
}
