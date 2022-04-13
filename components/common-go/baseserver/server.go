package baseserver

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	// TODO(milanpavlik): Move these to Server configuration options
	httpListenPort      = 9000
	defaultCloseTimeout = 3 * time.Second
)

type Server struct {
	// Name is the name of this Server.
	// Name is used for observability purposes to identify the server in telemetry data.
	Name string

	logger *logrus.Entry

	// httpListener is the underlying socket listener for the HTTP server
	httpListener net.Listener
	// httpMux is the HTTP router served by httpSrv
	// To register your handlers, use Server.HTTPMux() to obtain a reference
	httpMux *http.ServeMux
	// httpSrv is the underlying HTTP server we're managing
	httpSrv *http.Server
}

// ListenAndServe will start listeners to begin accepting traffic.
// ListenAndServe blocks until the servers either:
// 	* Fails to start
//	* Server.Close() is invoked and the server is terminated
// ListenAndServe will only return an error when ungraceful termination occurs - startup failure or the server termiantes for unexpected reasons.
// When ListenAndServer returns, all servers are terminated
func (s *Server) ListenAndServe() error {
	var err error

	s.httpListener, err = net.Listen("tcp", fmt.Sprintf(":%d", httpListenPort))
	if err != nil {
		return fmt.Errorf("failed to acquire port %d for http server: %w", httpListenPort, err)
	}

	errors := make(chan error)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger := s.Logger().WithField("protocol", "http")
		logger.Infof("Serving HTTP on %s.", s.httpListener.Addr().String())

		if serveErr := s.httpSrv.Serve(s.httpListener); err != nil {
			errors <- serveErr
		}

		logger.Info("HTTP server exited.")
	}()

	// await termination of all servers
	wg.Wait()
	// all servers have terminated, and they are the only producers into the errors channel so we can close it now.
	close(errors)

	for serveErr := range errors {

	}

	return nil
}

// Close will attempt to gracefully terminate running servers.
// Close() should be called at most once, multiple invocations have undefined behaviour.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultCloseTimeout)
	defer cancel()
	return s.close(ctx)
}

func (s *Server) HTTPMux() *http.ServeMux {
	return s.httpMux
}

func (s *Server) Logger() *logrus.Entry {
	return s.logger
}

func (s *Server) close(ctx context.Context) error {
	s.logger.Info("Closing servers.")
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to close HTTP server: %w", err)
	}

	// http.Server.Shutdown() also closes the underlying net.Listener, we are just releasing the reference.
	s.httpListener = nil
	s.logger.Info("Successfully terminated HTTP server")

	return nil
}

func (s *Server) initHTTP() error {
	mux := s.initHTTPMux()
	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", httpListenPort),
		Handler: mux,
	}

	return nil
}

func (s *Server) initHTTPMux() *http.ServeMux {
	mux := http.NewServeMux()
	return mux
}
