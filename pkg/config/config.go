package config

import (
	"reflect"

	"github.com/geliar/manopus/pkg/connector"
	"github.com/geliar/yaml"
)

func init() {
	yaml.DefaultMapType = reflect.TypeOf(map[string]interface{}{})
}

// Config contains structure of the Manopus manifest
type Config struct {
	//Connectors describe connectors structure
	Connectors map[string]connector.ConnectorConfig `yaml:"connectors"`
	//Env variables which represent env part of context data
	Env map[string]interface{} `yaml:"env"`
	//Sequences list of config of sequences
	Sequences []SequenceConfig `yaml:"sequences"`
}
