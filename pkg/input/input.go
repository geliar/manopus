package input

import "context"

type Driver interface {
	Name() string
	Type() string
	RegisterHandler(ctx context.Context, handler Handler)
	Stop(ctx context.Context)
}

type Handler func(ctx context.Context, event *Event)

type Event struct {
	Input string
	Type  string
	ID    string
	Data  map[string]interface{}
}
