package simnet

import (
	"context"
	"errors"
	"fmt"
	"net"
)

var (
	// ErrNetworkPartitioned is returned when a network is partitioned.
	ErrNetworkPartitioned = errors.New("simnet: network partitioned")

	// ErrDialFailed is returned when a dial fails.
	ErrDialFailed = errors.New("simnet: dial failed")
)

// Dialer is a net.Dialer that simulates network conditions.
type Dialer struct {
	dialer net.Dialer // Underlying dialer (can be customized)
	config *Config    // Network simulation configuration
}

// NewDialer creates a new simulated Dialer with the given configuration.
func NewDialer(cfg *Config) *Dialer {
	return &Dialer{
		config: cfg,
	}
}

// DialContext simulates dialing a network connection.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if d.config.isPartitioned(address) {
		return nil, fmt.Errorf("%w: unable to reach address: %s", ErrNetworkPartitioned, address)
	}

	conn, err := d.dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrDialFailed, err)
	}
	return wrapConn(conn, d.config), nil
}

// Dial simulates dialing without context.
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

// Helper method to check if an address is partitioned.
func (cfg *Config) isPartitioned(address string) bool {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	_, partitioned := cfg.PartitionedAddrs[address]
	return partitioned
}
