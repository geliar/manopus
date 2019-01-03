package main

import (
	"context"
	"flag"

	"github.com/geliar/manopus/pkg/config"

	"github.com/geliar/manopus/pkg/log"
)

func main() {
	ctx := log.Logger.WithContext(context.Background())
	flag.Parse()
	configFiles := flag.Args()
	config.InitConfig(ctx, configFiles)
	ch := make(chan struct{})
	<-ch
}
