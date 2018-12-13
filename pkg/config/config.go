package config

import (
	"github.com/geliar/manopus/pkg/connector"
	"github.com/geliar/manopus/pkg/sequencer"
)

// Config contains structure of the Manopus manifest
type Config struct {
	//Connectors describe connectors structure
	Connectors map[string]connector.ConnectorConfig `yaml:"connectors"`
	//Env variables which represent env part of context data
	Env map[string]interface{} `yaml:"env"`
	//Handlers list of handlers
	Handlers []sequencer.HandlerConfig `yaml:"handlers"`
}
