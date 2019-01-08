package config

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"github.com/geliar/manopus/pkg/connector"
	"github.com/geliar/manopus/pkg/http"
	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/sequencer"

	"github.com/geliar/yaml"
)

func init() {
	yaml.DefaultMapType = reflect.TypeOf(map[string]interface{}{})
}

// Config contains structure of the Manopus manifest
type Config struct {
	//ShutdownTimeout timeout of graceful shutdown
	ShutdownTimeout int `yaml:"shutdown_timeout"`
	//Connectors describe connectors structure
	Connectors map[string]connector.ConnectorConfig `yaml:"connectors"`
	//Sequencer config
	Sequencer sequencer.Sequencer `yaml:"sequencer"`
	//HTTP server config
	HTTP http.HTTPConfig
}

func InitConfig(ctx context.Context, configs []string) (*Config, *sequencer.Sequencer, *http.HTTPServer) {
	l := logger(ctx)
	var files []string
	if len(configs) == 0 {
		return nil, nil, nil
	}
	for _, name := range configs {
		err := filepath.Walk(name,
			func(path string, info os.FileInfo, err error) error {
				if info.Mode().IsRegular() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml") {
					files = append(files, path)
				}
				return nil
			})
		if err != nil {
			l.Error().Err(err).Msgf("Cannot walk through path %s", name)
			continue
		}
	}

	var configBuffer []byte
	for _, f := range files {
		log.Info().Str("file", f).Msg("Reading config file")
		buf, _ := ioutil.ReadFile(f)
		configBuffer = append(configBuffer, buf...)
		configBuffer = append(configBuffer, []byte("\n")...)
	}

	var c Config
	if err := yaml.Unmarshal(configBuffer, &c); err != nil {
		l.Fatal().Err(err).Msg("Cannot parse config files")
	}
	for i := range c.Connectors {
		connector.Configure(ctx, i, c.Connectors[i])
	}

	c.Sequencer.Init(ctx)
	input.RegisterHandlerAll(ctx, c.Sequencer.Roll)
	h := http.Init(ctx, c.HTTP)
	return &c, &c.Sequencer, h
}
