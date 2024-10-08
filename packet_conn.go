package simnet

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

// simulatedPacketConn is a net.PacketConn that simulates network conditions
// such as latency, loss, duplication, and reordering.
type simulatedPacketConn struct {
	conn       net.PacketConn
	cfg        *Config
	localAddr  net.Addr
	remoteAddr net.Addr
	closed     chan struct{}
	readQueue  chan packet
	writeQueue chan packet
	rand       *rand.Rand
}

// packet represents a UDP packet, including the data and the address
// it was sent from or to (depending on whether it is incoming or outgoing).
type packet struct {
	data []byte
	addr net.Addr
}

// newSimulatedPacketConn creates a new simulatedPacketConn with the given
// underlying connection and network configuration.
func newSimulatedPacketConn(conn net.PacketConn, cfg *Config, rand *rand.Rand) *simulatedPacketConn {
	spc := &simulatedPacketConn{
		conn:       conn,
		cfg:        cfg,
		closed:     make(chan struct{}),
		readQueue:  make(chan packet, 100),
		writeQueue: make(chan packet, 100),
		rand:       rand,
	}

	// Start the read and write loops in separate goroutines.
	go spc.readLoop()
	go spc.writeLoop()

	return spc
}

// ReadFrom reads a packet from the connection, applying network conditions.
func (spc *simulatedPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	select {
	case pkt := <-spc.readQueue:
		n = copy(p, pkt.data)
		addr = pkt.addr
		return n, addr, nil
	case <-spc.closed:
		return 0, nil, net.ErrClosed
	}
}

// WriteTo writes a packet to the connection, applying network conditions.
func (spc *simulatedPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	if spc.cfg.isPartitioned(addr.String()) {
		return 0, fmt.Errorf("%w: unable to reach address: %s", ErrNetworkPartitioned, addr)
	}

	spc.enqueuePacket(packet{data: append([]byte(nil), p...), addr: addr})
	return len(p), nil
}

// Close closes the connection.
func (spc *simulatedPacketConn) Close() error {
	close(spc.closed)
	return spc.conn.Close()
}

// LocalAddr returns the local network address.
func (spc *simulatedPacketConn) LocalAddr() net.Addr {
	return spc.conn.LocalAddr()
}

// SetDeadline sets the read and write deadlines.
func (spc *simulatedPacketConn) SetDeadline(t time.Time) error {
	return spc.conn.SetDeadline(t)
}

// SetReadDeadline sets the read deadline.
func (spc *simulatedPacketConn) SetReadDeadline(t time.Time) error {
	return spc.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline.
func (spc *simulatedPacketConn) SetWriteDeadline(t time.Time) error {
	return spc.conn.SetWriteDeadline(t)
}

// readLoop reads packets from the underlying connection and enqueues them
// to be processed with network conditions applied.
func (spc *simulatedPacketConn) readLoop() {
	for {
		select {
		case <-spc.closed:
			return
		default:
			buf := make([]byte, 65535) // Maximum UDP packet size (64 KiB)
			n, addr, err := spc.conn.ReadFrom(buf)
			if err != nil {
				continue
			}

			pkt := packet{
				data: buf[:n],
				addr: addr,
			}
			spc.processIncomingPacket(pkt)
		}
	}
}

// writeLoop writes packets to the underlying connection with network conditions applied.
func (spc *simulatedPacketConn) writeLoop() {
	for {
		select {
		case <-spc.closed:
			return
		case pkt := <-spc.writeQueue:
			spc.processOutgoingPacket(pkt)
		}
	}
}

// enqueuePacket enqueues a packet to be processed with network conditions applied.
func (spc *simulatedPacketConn) enqueuePacket(pkt packet) {
	spc.cfg.mu.Lock()
	defer spc.cfg.mu.Unlock()

	// Simulate loss
	if spc.simulateLoss() {
		return // Drop the packet
	}

	// Simulate duplication
	if spc.simulateDuplication() {
		spc.deliverPacket(pkt)
	}

	// Simulate reordering
	if spc.simulateReordering() {
		go func() {
			time.Sleep(spc.simulateLatency(len(pkt.data)))
			spc.deliverPacket(pkt)
		}()
	} else {
		spc.deliverPacket(pkt)
	}
}

// deliverPacket delivers a packet to the read queue after applying network conditions.
func (spc *simulatedPacketConn) deliverPacket(pkt packet) {
	time.Sleep(spc.simulateLatency(len(pkt.data)))
	select {
	case spc.readQueue <- pkt:
	case <-spc.closed:
	}
}

// processIncomingPacket processes an incoming packet with network conditions applied.
func (spc *simulatedPacketConn) processIncomingPacket(pkt packet) {
	spc.enqueuePacket(pkt)
}

// processOutgoingPacket processes an outgoing packet with network conditions applied.
func (spc *simulatedPacketConn) processOutgoingPacket(pkt packet) {
	// Simulate sending the packet
	_, err := spc.conn.WriteTo(pkt.data, pkt.addr)
	if err != nil {
		// Handle error?
	}
}

// simulateLatency simulates network latency based on the configuration.
func (spc *simulatedPacketConn) simulateLatency(n int) time.Duration {
	latency := spc.cfg.Latency
	if spc.cfg.Jitter > 0 {
		jitter := time.Duration(spc.rand.Int63n(int64(spc.cfg.Jitter)))
		latency += jitter
	}
	if spc.cfg.Bandwidth > 0 && n > 0 {
		transferTime := time.Duration(float64(n) / float64(spc.cfg.Bandwidth) * float64(time.Second))
		latency += transferTime
	}
	return latency
}

// simulateLoss determines if a packet should be dropped based on the loss rate.
func (spc *simulatedPacketConn) simulateLoss() bool {
	return spc.cfg.LossRate > 0 && spc.rand.Float64() < spc.cfg.LossRate
}

// simulateReordering determines if a packet should be reordered based on the reorder rate.
func (spc *simulatedPacketConn) simulateReordering() bool {
	return spc.cfg.ReorderRate > 0 && spc.rand.Float64() < spc.cfg.ReorderRate
}

// simulateDuplication determines if a packet should be duplicated based on the duplicate rate.
func (spc *simulatedPacketConn) simulateDuplication() bool {
	return spc.cfg.DuplicateRate > 0 && spc.rand.Float64() < spc.cfg.DuplicateRate
}

// UDPConn creates a simulated UDP connection.
func UDPConn(cfg *Config, laddr, raddr *net.UDPAddr) (net.PacketConn, error) {
	if cfg == nil {
		cfg = NewConfig()
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}

	rand := cfg.randSource()
	spc := newSimulatedPacketConn(conn, cfg, rand)
	return spc, nil
}
