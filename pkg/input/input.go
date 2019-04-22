package input

import (
	"context"

	"github.com/geliar/manopus/pkg/payload"
)

// Driver input driver interface
type Driver interface {
	// Name returns name of the driver
	Name() string
	// Type returns type of the driver
	Type() string
	// RegisterHandler register handler with in the driver
	RegisterHandler(ctx context.Context, handler Handler)
	// Stop stop the driver
	Stop(ctx context.Context)
}

// Handler describes input event handler function
type Handler func(ctx context.Context, event *payload.Event) (callback interface{})
