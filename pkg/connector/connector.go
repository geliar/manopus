package connector

type ConnectorConfig struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

type Builder func(name string, config map[string]interface{})
