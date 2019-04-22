package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/geliar/manopus/pkg/store"

	"github.com/geliar/manopus/pkg/config"
	"github.com/geliar/manopus/pkg/http"
	"github.com/geliar/manopus/pkg/input"
	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/output"
	"github.com/geliar/manopus/pkg/sequencer"

	flag "github.com/ogier/pflag"
)

var help = flag.BoolP("help", "h", false, "Show this page")
var noload = flag.BoolP("noload", "n", false, "Don't load unfinished sequences from store")

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = log.Logger.WithContext(ctx)

	flag.Parse()
	configFiles := flag.Args()

	if *help || len(configFiles) == 0 {
		showUsage()
		return
	}

	log.Info().Msg("Starting Manopus...")

	cfg, sequencerInstance, httpServer := config.InitConfig(ctx, configFiles, *noload)
	if cfg == nil || sequencerInstance == nil {
		log.Fatal().Msg("No configuration provided")
	}
	wait(ctx, cancel, cfg, sequencerInstance, httpServer)
}

func showUsage() {
	println("Usage: " + os.Args[0] + " [options] [config files or dirs]...")
	println("Starts Manopus omnichannel automation bot\n")
	println("Options and flags:")
	println("  -n, --noload: Don't load unfinished sequences from store")
	println("  -h, --help: Show this page")
}

func wait(ctx context.Context, cancel context.CancelFunc, cfg *config.Config, sequencerInstance *sequencer.Sequencer, httpServer *http.Server) {
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
			store.StopAll(ctx)
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
