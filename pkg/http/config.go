package http

type HTTPConfig struct {
	Listen          string `yaml:"listen"`
	ShutdownTimeout int    `yaml:"shutdown_timeout"`
}
