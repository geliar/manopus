package input

import (
	"context"

	"github.com/geliar/manopus/pkg/payload"
)

type Driver interface {
	Name() string
	Type() string
	RegisterHandler(ctx context.Context, handler Handler)
	Stop(ctx context.Context)
}

type Handler func(ctx context.Context, event *payload.Event) (callback interface{})
