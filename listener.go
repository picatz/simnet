package simnet

import (
	"errors"
	"fmt"
	"net"
)

// ErrFailedToAccept is returned when a connection cannot be accepted.
var ErrFailedToAccept = errors.New("simnet: failed to accept connection")

// Listener is a net.Listener that simulates network conditions.
type Listener struct {
	ln  net.Listener
	cfg *Config
}

// NewListener wraps an existing net.Listener with simulated network conditions.
func NewListener(ln net.Listener, cfg *Config) net.Listener {
	return &Listener{
		ln:  ln,
		cfg: cfg,
	}
}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.ln.Accept()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFailedToAccept, err)
	}
	// Wrap the connection with simulated network conditions.
	return wrapConn(conn, l.cfg), nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *Listener) Close() error {
	return l.ln.Close()
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.ln.Addr()
}
