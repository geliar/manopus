package connector

import "context"

// Config connector configuration structure
type Config struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

// Builder connector builder description
type Builder func(ctx context.Context, name string, config map[string]interface{})
