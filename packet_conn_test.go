package simnet_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/picatz/simnet"
	"github.com/shoenig/test/must"
	"github.com/shoenig/test/portal"
)

func ExampleUDPConn() {
	cfg := &simnet.Config{
		// Latency:       100 * time.Millisecond,
		// Jitter:        50 * time.Millisecond,
		// Bandwidth:     256 * 1024, // 256 KBps
		// LossRate:      0.05,       // 5% packet loss
		// ReorderRate:   0.05,       // 5% packet reordering
		// DuplicateRate: 0.05,       // 5% packet duplication
		// Seed:          42,         // Seed for deterministic behavior
	}

	localAddr := &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 8080,
	}

	remoteAddr := &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 8081,
	}

	conn, err := simnet.UDPConn(cfg, localAddr, remoteAddr)
	if err != nil {
		panic(err)
	}

	_, err = conn.WriteTo([]byte("Hello, simnet!"), remoteAddr)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1024)
	n, addr, err := conn.ReadFrom(buf)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(buf[:n]), addr)

	// Output:
	// Hello, simnet! 127.0.0.1:8081
}

func TestUDPConn(t *testing.T) {
	checkConnBascis := func(t *testing.T, conn net.PacketConn) {
		_, err := conn.WriteTo([]byte("Hello, simnet!"), &net.UDPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 8081,
		})
		must.NoError(t, err)

		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFrom(buf)
		must.NoError(t, err)
		must.Eq(t, n, 14)
		must.Eq(t, addr.String(), "127.0.0.1:8081")
	}

	tests := []struct {
		name  string
		cfg   *simnet.Config
		check func(*testing.T, net.PacketConn)
	}{
		{
			name:  "no network config",
			cfg:   nil,
			check: checkConnBascis,
		},
		{
			name:  "with empty network config",
			cfg:   &simnet.Config{},
			check: checkConnBascis,
		},
		{
			name: "with latency",
			cfg: &simnet.Config{
				Latency: 100, // 100ms
			},
			check: checkConnBascis,
		},
		{
			name: "with jitter",
			cfg: &simnet.Config{
				Jitter: 50, // 50ms
			},
			check: checkConnBascis,
		},
		{
			name: "with bandwidth",
			cfg: &simnet.Config{
				Bandwidth: 256 * 1024, // 256 KBps
			},
			check: checkConnBascis,
		},
		{
			name: "with loss rate",
			cfg: &simnet.Config{
				LossRate: 0.05, // 5% packet loss
			},
			check: checkConnBascis,
		},
		{
			name: "with reorder rate",
			cfg: &simnet.Config{
				ReorderRate: 0.05, // 5% packet reordering
			},
			check: checkConnBascis,
		},
		{
			name: "with duplicate rate",
			cfg: &simnet.Config{
				DuplicateRate: 0.05, // 5% packet duplication
			},
			check: checkConnBascis,
		},
		{
			name: "with seed",
			cfg: &simnet.Config{
				Seed: 42, // Seed for deterministic behavior
			},
			check: checkConnBascis,
		},
		{
			name: "with all network config",
			cfg: &simnet.Config{
				Latency:       100,        // 100ms
				Jitter:        50,         // 50ms
				Bandwidth:     256 * 1024, // 256 KBps
				LossRate:      0.05,       // 5% packet loss
				ReorderRate:   0.05,       // 5% packet reordering
				DuplicateRate: 0.05,       // 5% packet duplication
				Seed:          42,         // Seed for deterministic behavior
			},
			check: checkConnBascis,
		},
	}

	g := portal.New(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ports := g.Grab(2)

			localAddr := &net.UDPAddr{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: ports[0],
			}

			remoteAddr := &net.UDPAddr{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: ports[1],
			}

			conn, err := simnet.UDPConn(test.cfg, localAddr, remoteAddr)
			must.NoError(t, err)
			t.Cleanup(func() {
				err := conn.Close()
				must.NoError(t, err)
			})

			test.check(t, conn)
		})
	}
}
