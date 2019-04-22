package output

import (
	"context"

	"github.com/geliar/manopus/pkg/payload"
)

// Driver output driver interface
type Driver interface {
	// Name returns name of the output
	Name() string
	// Type returns type of output
	Type() string
	// Send sends response with output
	Send(ctx context.Context, response *payload.Response) map[string]interface{}
	//Stop stops execution of output
	Stop(ctx context.Context)
}
