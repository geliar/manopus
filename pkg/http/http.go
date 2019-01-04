package http

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/geliar/manopus/pkg/log"
)

type HTTPServer struct {
	config   HTTPConfig
	instance *http.Server
	routes   map[string]http.Handler
	sync.RWMutex
}

func Init(ctx context.Context, config HTTPConfig) *HTTPServer {
	if config.Listen == "" {
		log.Info().Msg("HTTP server has not been configured")
		return nil
	}

	server := new(HTTPServer)
	server.config = config
	server.Start(ctx)
	return server
}

func (s *HTTPServer) Start(ctx context.Context) {
	l := logger(ctx)
	s.instance = &http.Server{
		Addr:    s.config.Listen,
		Handler: http.HandlerFunc(s.routerHandler),
	}
	go func() {
		l.Info().Msgf("Listening HTTP requests on %s", s.config.Listen)
		err := s.instance.ListenAndServe()
		if err != http.ErrServerClosed {
			l.Fatal().Err(err).Msg("Error on HTTP server listener")
		}
	}()
}

func (s *HTTPServer) Stop(ctx context.Context) {
	l := logger(ctx)
	if s.instance == nil {
		log.Fatal().Msg("Trying to shutdown not started HTTP server")
	}
	l.Info().Msg("Shutting down HTTP server")
	if s.config.ShutdownTimeout != 0 {
		ctx, _ = context.WithTimeout(ctx, time.Duration(s.config.ShutdownTimeout)*time.Second)
	}
	err := s.instance.Shutdown(ctx)
	if err != nil {
		l.Error().Err(err).Msg("Didn't manage to shutdown HTTP server gracefully")
	}
}

func (s *HTTPServer) routerHandler(w http.ResponseWriter, r *http.Request) {
	s.RLock()
	defer s.RUnlock()
	for k, v := range s.routes {
		if strings.Contains(r.RequestURI, k) {
			v.ServeHTTP(w, r)
			return
		}
	}
}
