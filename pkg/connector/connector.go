package connector

import "context"

type Config struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

type Builder func(ctx context.Context, name string, config map[string]interface{})
