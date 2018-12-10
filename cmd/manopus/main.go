package main

import (
	"fmt"

	"io/ioutil"

	"github.com/DLag/manopus/pkg/connector"
	"github.com/DLag/manopus/pkg/log"
	"gopkg.in/yaml.v2"
)

func main() {
	files, err := ioutil.ReadDir("./examples/dialog/")
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot read dir")
	}
	var configBuffer []byte
	for _, f := range files {
		fmt.Println(f.Name())
		buf, _ := ioutil.ReadFile("./examples/dialog/" + f.Name())
		configBuffer = append(configBuffer, buf...)
		configBuffer = append(configBuffer, []byte("\n")...)
	}
	println(string(configBuffer))
	var config Config
	fmt.Println(yaml.Unmarshal(configBuffer, &config))
	for i := range config.Connectors {
		println(i)
		connector.Configure(i, config.Connectors[i])
	}
}
