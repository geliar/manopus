package output

import (
	"context"
	"sync"

	"github.com/geliar/manopus/pkg/payload"
)

type catalogStore struct {
	outputs map[string]Driver
	sync.RWMutex
}

var catalog catalogStore

// Register register connector in the catalog
func Register(ctx context.Context, name string, driver Driver) {
	catalog.register(ctx, name, driver)
}

// Send send response with specified connector
func Send(ctx context.Context, response *payload.Response) map[string]interface{} {
	return catalog.send(ctx, response)
}

// StopAll stop all outputs
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

func (c *catalogStore) send(ctx context.Context, response *payload.Response) map[string]interface{} {
	c.RLock()
	l := logger(ctx).With().Str("output_driver_name", response.Output).Logger()
	if _, ok := c.outputs[response.Output]; !ok {
		c.RUnlock()
		l.Error().
			Msgf("Cannot find output driver with name '%s'", response.Output)
		return nil
	}
	o := c.outputs[response.Output]
	c.RUnlock()
	return o.Send(l.WithContext(ctx), response)
}

func (c *catalogStore) stopAll(ctx context.Context) {
	c.Lock()
	defer c.Unlock()
	l := logger(ctx)

	for i := range c.outputs {
		l.Info().
			Str("output_driver_name", c.outputs[i].Name()).
			Str("output_driver_type", c.outputs[i].Type()).
			Msgf("Shutting down output driver")
		c.outputs[i].Stop(ctx)
		delete(c.outputs, i)
	}
}
