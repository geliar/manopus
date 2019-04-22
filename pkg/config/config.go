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
	"github.com/geliar/manopus/pkg/report"
	"github.com/geliar/manopus/pkg/sequencer"
	"github.com/geliar/manopus/pkg/store"

	"github.com/geliar/yaml"
)

func init() {
	yaml.DefaultMapType = reflect.TypeOf(map[string]interface{}{})
}

// Config contains structure of the Manopus manifest
type Config struct {
	//ShutdownTimeout timeout of graceful shutdown
	ShutdownTimeout int `yaml:"shutdown_timeout"`
	//Stores describe stores structure
	Stores map[string]store.Config `yaml:"stores"`
	//Connectors describe connectors structure
	Connectors map[string]connector.Config `yaml:"connectors"`
	//Sequencer config
	Sequencer sequencer.Sequencer `yaml:"sequencer"`
	//Report config
	Report report.Config
	//HTTP server config
	HTTP http.Config
}

// InitConfig initializes Manopus with configuration data
func InitConfig(ctx context.Context, configs []string, noload bool) (*Config, *sequencer.Sequencer, *http.Server) {
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
	log.Info().Strs("files", files).Msg("Reading config files")
	for _, f := range files {
		buf, _ := ioutil.ReadFile(f)
		configBuffer = append(configBuffer, buf...)
		configBuffer = append(configBuffer, []byte("\n")...)
	}

	var c Config

	if err := yaml.Unmarshal(configBuffer, &c); err != nil {
		l.Fatal().Err(err).Msg("Cannot parse config files")
	}

	//HTTP server
	h := http.Init(ctx, c.HTTP)

	//Connectors
	for i := range c.Connectors {
		connector.Configure(ctx, i, c.Connectors[i])
	}

	//Stores
	for i := range c.Stores {
		store.ConfigureStore(ctx, i, c.Stores[i])
	}

	//Report
	report.Init(ctx, c.Report)

	//Sequencer
	c.Sequencer.Init(ctx, noload)
	input.RegisterHandlerAll(ctx, c.Sequencer.Roll)
	l.Info().Msg("Configuration stage is complete")
	return &c, &c.Sequencer, h
}
