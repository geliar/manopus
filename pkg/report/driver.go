package report

import (
	"context"
	"io"
)

type Builder func(config map[string]interface{}, id string, step int) Driver

type Driver interface {
	Type() string
	PushString(ctx context.Context, report string)
	PushReader(ctx context.Context, report io.Reader)
	Close(ctx context.Context)
}
