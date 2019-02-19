package http

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/rs/zerolog/hlog"

	"github.com/DLag/midsimple"

	"github.com/geliar/manopus/pkg/log"
)

type HTTPServer struct {
	config       HTTPConfig
	instance     *http.Server
	routes       map[string]http.Handler
	defaultRoute http.Handler
	sync.RWMutex
	mainCtx context.Context
}

var server HTTPServer

func Init(ctx context.Context, config HTTPConfig) *HTTPServer {
	if config.Listen == "" {
		log.Info().Msg("HTTP server has not been configured")
		return nil
	}

	server.config = config
	server.mainCtx = ctx
	server.Start(ctx)
	return &server
}

func AddHandler(ctx context.Context, path string, h http.Handler) {
	server.AddHandler(ctx, path, h)
}

func SetDefaultHandler(ctx context.Context, h http.Handler) {
	server.SetDefaultHandler(ctx, h)
}

func (s *HTTPServer) Start(ctx context.Context) {
	l := logger(ctx)
	handler := midsimple.New(hlog.NewHandler(l)).
		Use(hlog.RemoteAddrHandler("ip")).
		Use(hlog.UserAgentHandler("user_agent")).
		Use(hlog.RefererHandler("referer")).
		WrapFunc(s.routerHandler)
	s.instance = &http.Server{
		Addr:    s.config.Listen,
		Handler: handler,
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
	err := s.instance.Shutdown(ctx)
	if err != nil {
		l.Error().Err(err).Msg("Didn't manage to shutdown HTTP server gracefully")
	}
}

func (s *HTTPServer) AddHandler(ctx context.Context, path string, h http.Handler) {
	l := logger(ctx).With().Str("http_path", path).Logger()
	s.Lock()
	defer s.Unlock()
	if s.routes == nil {
		s.routes = map[string]http.Handler{}
	}
	if _, ok := s.routes[path]; ok {
		l.Error().Msg("Trying to add HTTP handler for existing route")
		return
	}
	s.routes[path] = h
	l.Debug().Msg("Added HTTP server handler")
}

func (s *HTTPServer) SetDefaultHandler(ctx context.Context, h http.Handler) {
	l := logger(ctx)
	s.Lock()
	defer s.Unlock()
	s.defaultRoute = h
	l.Debug().Msg("Set default HTTP server handler")
}

func (s *HTTPServer) routerHandler(w http.ResponseWriter, r *http.Request) {
	hlog.FromRequest(r).Debug().
		Msgf("%s %s", r.Method, r.RequestURI)
	s.RLock()
	defer s.RUnlock()
	for k, v := range s.routes {
		if strings.HasPrefix(r.RequestURI, k) {
			v.ServeHTTP(w, r)
			return
		}
	}
	if s.defaultRoute != nil {
		s.defaultRoute.ServeHTTP(w, r)
	}
}
