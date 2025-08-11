package report

// Config report configuration structure
type Config struct {
	Driver string                 `yaml:"driver"`
	Config map[string]interface{} `yaml:"config"`
}
