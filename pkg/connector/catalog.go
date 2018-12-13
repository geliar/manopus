package connector

import (
	"sync"
)

type catalogStore struct {
	connectors map[string]Builder
	sync.RWMutex
}

var catalog catalogStore

func Register(name string, driver Builder) {
	catalog.register(name, driver)
}

func Configure(name string, connector ConnectorConfig) {
	catalog.configure(name, connector)
}

func (c *catalogStore) register(name string, driver Builder) {
	c.Lock()
	defer c.Unlock()
	l := logger()
	if c.connectors == nil {
		l.Debug().
			Msg("Initializing connector catalog")
		c.connectors = make(map[string]Builder)
	}
	l = l.With().Str("controller_name", name).Logger()
	if _, ok := c.connectors[name]; ok {
		l.Fatal().
			Msg("Trying to register connector with existing name")
	}
	c.connectors[name] = driver
	l.Info().
		Msg("Registered new connector")
}

func (c *catalogStore) configure(name string, connector ConnectorConfig) {
	c.RLock()
	defer c.RUnlock()
	l := logger()
	l = l.With().Str("controller_name", name).Logger()
	if _, ok := c.connectors[connector.Type]; ok {
		c.connectors[connector.Type](name, connector.Config)
	} else {
		l.Error().Msgf("Cannot find connector with type %q", connector.Type)
	}
}
