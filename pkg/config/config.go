package config

import (
	"reflect"

	"github.com/geliar/manopus/pkg/sequencer"

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
	//Sequencer config
	Sequencer sequencer.Sequencer `yaml:"sequencer"`
}
