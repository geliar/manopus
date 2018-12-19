package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/geliar/manopus/pkg/config"
	"github.com/geliar/manopus/pkg/connector"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/yaml"
)

func main() {
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
	fmt.Println(yaml.Unmarshal(configBuffer, &c))
	for i := range c.Connectors {
		//println(i)
		connector.Configure(context.Background(), i, c.Connectors[i])
	}
}
