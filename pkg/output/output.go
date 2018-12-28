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
	Encoding    string `yaml:"encoding"`
	Type        string `yaml:"type"`
}

type Response struct {
	// ID of original request
	ID string
	// Encoding defines should be response data decoded or not (plain, json, csv, tsv)
	Encoding string
	// Type of response. Default value depends on connector defaults.
	Type string
	// Data response data
	Data interface{}
	// Request original request
	Request *input.Event
}
