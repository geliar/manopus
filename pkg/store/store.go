package store

import "context"

// Store Manopus storage interface
type Store interface {
	//Name get name of the Store
	Name() string
	//Type get type of the Store
	Type() string
	//Save saves value with provided key
	Save(ctx context.Context, key string, value []byte) (err error)
	//Load loads value with provided key
	Load(ctx context.Context, key string) (value []byte, err error)
	//Stop stops store instance
	Stop(ctx context.Context)
}
