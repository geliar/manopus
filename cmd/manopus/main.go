package main

import (
	"context"
	"io/ioutil"

	"github.com/geliar/manopus/pkg/input"

	"github.com/geliar/manopus/pkg/sequencer"

	"github.com/geliar/manopus/pkg/config"
	"github.com/geliar/manopus/pkg/connector"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/yaml"
)

func main() {
	ctx := log.Logger.WithContext(context.Background())
	l := log.Ctx(ctx)
	files, err := ioutil.ReadDir("./examples/dialog/")
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot read dir")
	}
	var configBuffer []byte
	for _, f := range files {
		//fmt.Println(f.Name())
		buf, _ := ioutil.ReadFile("./examples/dialog/" + f.Name())
		configBuffer = append(configBuffer, buf...)
		configBuffer = append(configBuffer, []byte("\n")...)
	}
	//println(string(configBuffer))
	var c config.Config
	if err := yaml.Unmarshal(configBuffer, &c); err != nil {
		l.Fatal().Err(err).Msg("Cannot read config files")
	}
	for i := range c.Connectors {
		connector.Configure(ctx, i, c.Connectors[i])
	}
	s := sequencer.Sequencer{}
	input.RegisterHandlerAll(ctx, s.Roll)
	ch := make(chan struct{})
	<-ch
}
