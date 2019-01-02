package output

import (
	"context"

	"github.com/geliar/manopus/pkg/input"
)

type Driver interface {
	Name() string
	Type() string
	Send(ctx context.Context, response *Response)
	Stop(ctx context.Context)
}

type OutputConfig struct {
	Destination string `yaml:"destination"`
	Type        string `yaml:"type"`
}

type Response struct {
	// ID of original request
	ID string
	// Type of response. Default value depends on connector defaults.
	Type string
	// Data response data
	Data map[string]interface{}
	// Encoding of response (none, plain, json, json, toml)
	Encoding string
	// Request original request
	Request *input.Event
}
