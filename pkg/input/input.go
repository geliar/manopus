package input

import "context"

type Driver interface {
	Name() string
	Type() string
	RegisterHandler(handler Handler)
	Stop(ctx context.Context)
}

type Handler func(event Event)

type Event struct {
	Input string
	Type  string
	ID    string
	Data  map[string]interface{}
}
