package httptest

import (
	"net/http"
	"net/http/httptest"

	"github.com/picatz/simnet"
)

// Server is a simulated HTTP server that applies network conditions.
type Server struct {
	srv *httptest.Server
}

// NewServer starts and returns a new simulated HTTP server.
func NewServer(cfg *simnet.Config, handler http.Handler) *Server {
	s := &Server{}
	s.srv = httptest.NewUnstartedServer(handler)
	s.wrapListener(cfg)
	s.srv.Start()
	return s
}

// NewTLSServer starts and returns a new simulated HTTPS server using TLS.
func NewTLSServer(cfg *simnet.Config, handler http.Handler) *Server {
	s := &Server{}
	s.srv = httptest.NewUnstartedServer(handler)
	s.wrapListener(cfg)
	s.srv.StartTLS()
	return s
}

// Close shuts down the server and blocks until all outstanding requests on this server have completed.
func (s *Server) Close() {
	s.srv.Close()
}

// wrapListener wraps the server's listener with a simulated listener.
func (s *Server) wrapListener(cfg *simnet.Config) {
	originalListener := s.srv.Listener
	s.srv.Listener = simnet.NewListener(originalListener, cfg)
}

// Client returns an HTTP client configured to make requests to the server.
func (s *Server) Client() *http.Client {
	return s.srv.Client()
}

// URL returns the base URL of the server.
func (s *Server) URL() string {
	return s.srv.URL
}
