package report

import (
	"context"
	"sync"
)

type catalogStore struct {
	builders map[string]Builder
	sync.RWMutex
}

var catalog catalogStore

// Register register report builder in the catalog
func Register(ctx context.Context, name string, builder Builder) {
	catalog.register(ctx, name, builder)
}

// Open start the report
func Open(ctx context.Context, id string, step int) Driver {
	return catalog.open(ctx, id, step)
}

func (c *catalogStore) register(ctx context.Context, name string, builder Builder) {
	c.Lock()
	defer c.Unlock()
	l := logger(ctx)
	if c.builders == nil {
		l.Debug().
			Msg("Initializing report drivers catalog")
		c.builders = make(map[string]Builder)
	}
	l = l.With().
		Str("report_name", name).
		Logger()
	if _, ok := c.builders[name]; ok {
		l.Fatal().
			Msg("Cannot register driver with existing name")
	}
	c.builders[name] = builder
	l.Debug().
		Msg("Registered new driver")
}

func (c *catalogStore) open(ctx context.Context, id string, step int) Driver {
	c.RLock()

	l := logger(ctx)
	if _, ok := c.builders[config.Driver]; !ok {
		c.RUnlock()
		l.Error().
			Str("report_name", config.Driver).
			Msgf("Cannot find driver with name '%s'", config.Driver)
		return nil
	}
	d := c.builders[config.Driver](config.Config, id, step)
	c.RUnlock()
	return d
}
