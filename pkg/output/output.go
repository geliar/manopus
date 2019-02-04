package output

import (
	"context"

	"github.com/geliar/manopus/pkg/payload"
)

type Driver interface {
	Name() string
	Type() string
	Send(ctx context.Context, response *payload.Response) map[string]interface{}
	Stop(ctx context.Context)
}
