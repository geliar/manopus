package report

import (
	"context"
	"io"
)

// Builder report driver builder
type Builder func(config map[string]interface{}, id string, step int) Driver

// Driver report driver interface
type Driver interface {
	// Type returns type of the driver
	Type() string
	// PushString pushes string to report
	PushString(ctx context.Context, report string)
	// PushReader pushes io.Reader to report to allow streaming of log to report
	PushReader(ctx context.Context, report io.Reader)
	// Close close the report
	Close(ctx context.Context)
}
