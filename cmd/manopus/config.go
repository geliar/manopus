package main

import "github.com/DLag/manopus/pkg/connector"

type Config struct {
	Connectors map[string]connector.ConnectorConfig `yaml:"connectors"`
	Vars map[string]interface{} `yaml:"vars"`
}