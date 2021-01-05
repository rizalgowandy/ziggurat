package server

import (
	"context"
	logger2 "github.com/gojekfarm/ziggurat/logger"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var defaultHTTPPort = "8080"

type DefaultHttpServer struct {
	server *http.Server
	router *httprouter.Router
}

func WithPort(port string) func(s *DefaultHttpServer) {
	return func(s *DefaultHttpServer) {
		s.server.Addr = "localhost:" + port
	}
}

func NewHTTPServer(opts ...func(s *DefaultHttpServer)) *DefaultHttpServer {
	router := httprouter.New()
	logger := logger2.NewJSONLogger("info")
	httpRequestLogger := HTTPRequestLogger(logger)
	server := &http.Server{Handler: httpRequestLogger(router)}
	s := &DefaultHttpServer{
		server: server,
		router: router,
	}
	for _, opt := range opts {
		opt(s)
	}
	s.server.Addr = "localhost:" + defaultHTTPPort
	return s
}

func (s *DefaultHttpServer) Run(ctx context.Context) chan error {
	errorChan := make(chan error)
	s.router.GET("/v1/ping", pingHandler)
	go func(server *http.Server) {
		if err := server.ListenAndServe(); err != nil {
			errorChan <- err
		}
	}(s.server)

	go func() {
		done := ctx.Done()
		<-done
		if err := s.server.Shutdown(ctx); err != nil {
			errorChan <- err
		}
	}()

	return errorChan
}

func (s *DefaultHttpServer) ConfigureHTTPEndpoints(f func(r *httprouter.Router)) {
	f(s.router)
}

func (s *DefaultHttpServer) ConfigureHandler(f func(r *httprouter.Router) http.Handler) {
	s.server.Handler = f(s.router)
}
