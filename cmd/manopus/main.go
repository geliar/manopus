package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/geliar/manopus/pkg/config"
	"github.com/geliar/manopus/pkg/http"
	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/output"
	"github.com/geliar/manopus/pkg/sequencer"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = log.Logger.WithContext(ctx)

	log.Info().Msg("Starting Manopus...")
	flag.Parse()
	configFiles := flag.Args()
	cfg, sequencerInstance, httpServer := config.InitConfig(ctx, configFiles)
	if cfg == nil || sequencerInstance == nil {
		log.Fatal().Msg("No configuration provided")
	}
	wait(ctx, cancel, cfg, sequencerInstance, httpServer)
}

func wait(ctx context.Context, cancel context.CancelFunc, cfg *config.Config, sequencerInstance *sequencer.Sequencer, httpServer *http.HTTPServer) {
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt)
	for {
		select {
		case <-stopSignal:
			log.Info().Msg("Interrupt signal received")
			if cfg.ShutdownTimeout != 0 {
				log.Info().Msgf("%d seconds for graceful shutdown", cfg.ShutdownTimeout)
				var shutCancel context.CancelFunc
				timeout := time.Duration(cfg.ShutdownTimeout) * time.Second
				ctx, shutCancel = context.WithTimeout(ctx, timeout)
				defer shutCancel()
				time.AfterFunc(timeout, func() {
					log.Info().Msgf("Forcing shutdown")
					cancel()
				})
			}
			sequencerInstance.Stop(ctx)
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				httpServer.Stop(ctx)
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				input.StopAll(ctx)
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				output.StopAll(ctx)
				wg.Done()
			}()
			wg.Wait()
			log.Info().Msg("Manopus has been gracefully stopped")
			return
		}
	}
}
