package connector

import (
	"context"
	"sync"
)

type catalogStore struct {
	connectors map[string]Builder
	sync.RWMutex
}

var catalog catalogStore

// Register register connector builder in the catalog
func Register(ctx context.Context, name string, driver Builder) {
	catalog.register(ctx, name, driver)
}

// Configure register specified connector with configuration and connector builder from the catalog
func Configure(ctx context.Context, name string, connector Config) {
	catalog.configure(ctx, name, connector)
}

func (c *catalogStore) register(ctx context.Context, name string, driver Builder) {
	c.Lock()
	defer c.Unlock()
	l := logger(ctx)
	if c.connectors == nil {
		l.Debug().
			Msg("Initializing connector catalog")
		c.connectors = make(map[string]Builder)
	}
	l = l.With().Str("connector_name", name).Logger()
	if _, ok := c.connectors[name]; ok {
		l.Fatal().
			Msg("Trying to register connector with existing name")
	}
	c.connectors[name] = driver
	l.Debug().
		Msg("Registered new connector")
}

func (c *catalogStore) configure(ctx context.Context, name string, connector Config) {
	c.RLock()
	defer c.RUnlock()
	l := logger(ctx)
	l = l.With().Str("connector_name", name).Logger()
	if _, ok := c.connectors[connector.Type]; ok {
		c.connectors[connector.Type](ctx, name, connector.Config)
	} else {
		l.Warn().Msgf("Cannot find connector with type '%s'", connector.Type)
	}
}
