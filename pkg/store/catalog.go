package store

import (
	"context"
	"sync"
)

type Config struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

type Builder func(ctx context.Context, name string, config map[string]interface{})

type catalogBuilders struct {
	builders map[string]Builder
	sync.RWMutex
}

type catalogStores struct {
	stores map[string]Store
	sync.RWMutex
}

var (
	stores   catalogStores
	builders catalogBuilders
)

func RegisterBuilder(ctx context.Context, name string, builder Builder) {
	builders.register(ctx, name, builder)
}

func RegisterStore(ctx context.Context, store Store) {
	stores.register(ctx, store)
}

func Save(ctx context.Context, name string, key string, value []byte) (err error) {
	return stores.save(ctx, name, key, value)
}

func Load(ctx context.Context, name string, key string) (value []byte, err error) {
	return stores.load(ctx, name, key)
}

func ConfigureStore(ctx context.Context, name string, store Config) {
	builders.configure(ctx, name, store)
}

func (c *catalogStores) register(ctx context.Context, store Store) {
	c.Lock()
	defer c.Unlock()
	l := logger(ctx)
	if c.stores == nil {
		l.Debug().
			Msg("Initializing store catalog")
		c.stores = make(map[string]Store)
	}
	l = l.With().
		Str("store_name", store.Name()).
		Logger()
	if _, ok := c.stores[store.Name()]; ok {
		l.Fatal().
			Msg("Cannot register store with existing name")
	}
	c.stores[store.Name()] = store
	l.Debug().
		Msg("Registered new store")
}

func StopAll(ctx context.Context) {
	stores.stopAll(ctx)
}

func (c *catalogStores) save(ctx context.Context, name string, key string, value []byte) (err error) {
	c.RLock()

	l := logger(ctx)
	if _, ok := c.stores[name]; !ok {
		c.RUnlock()
		l.Error().
			Str("store_name", name).
			Msgf("Cannot find store with name '%s'", name)
		return
	}
	p := c.stores[name]
	c.RUnlock()
	return p.Save(ctx, key, value)
}

func (c *catalogStores) load(ctx context.Context, name string, key string) (value []byte, err error) {
	c.RLock()

	l := logger(ctx)
	if _, ok := c.stores[name]; !ok {
		c.RUnlock()
		l.Error().
			Str("store_name", name).
			Msgf("Cannot find store with name '%s'", name)
		return
	}
	p := c.stores[name]
	c.RUnlock()
	return p.Load(ctx, key)
}

func (c *catalogStores) stopAll(ctx context.Context) {
	c.Lock()
	defer c.Unlock()
	l := logger(ctx)

	for i := range c.stores {
		l.Info().
			Str("store_name", c.stores[i].Name()).
			Str("store_type", c.stores[i].Type()).
			Msgf("Shutting down store")
		c.stores[i].Stop(ctx)
		delete(c.stores, i)
	}
}

func (c *catalogBuilders) configure(ctx context.Context, name string, store Config) {
	c.RLock()
	defer c.RUnlock()
	l := logger(ctx)
	l = l.With().Str("store_name", name).Logger()
	if _, ok := c.builders[store.Type]; ok {
		c.builders[store.Type](ctx, name, store.Config)
	} else {
		l.Warn().Msgf("Cannot find store builder with type '%s'", store.Type)
	}
}

func (c *catalogBuilders) register(ctx context.Context, name string, builder Builder) {
	c.Lock()
	defer c.Unlock()
	l := logger(ctx)
	if c.builders == nil {
		l.Debug().
			Msg("Initializing store builders catalog")
		c.builders = make(map[string]Builder)
	}
	l = l.With().
		Str("store_type", name).
		Logger()
	if _, ok := c.builders[name]; ok {
		l.Fatal().
			Msg("Cannot register store builder with existing name")
	}
	c.builders[name] = builder
	l.Debug().
		Msg("Registered new store builder")
}
