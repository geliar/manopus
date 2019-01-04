package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"

	"github.com/geliar/manopus/pkg/config"
	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/output"
)

func main() {
	ctx := log.Logger.WithContext(context.Background())
	log.Info().Msg("Starting Manopus...")
	flag.Parse()
	configFiles := flag.Args()
	sequencer, httpServer := config.InitConfig(ctx, configFiles)
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, os.Kill)
	for {
		select {
		case <-stopSignal:
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				httpServer.Stop(ctx)
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				sequencer.Stop(ctx)
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				input.StopAll(ctx)
				wg.Done()
			}()
			go func() {
				output.StopAll(ctx)
				wg.Done()
			}()
			wg.Wait()
			log.Info().Msg("Manopus has been gracefully stopped")
		}
	}
}
