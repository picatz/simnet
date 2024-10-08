package simnet

import (
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

// simulatedConn is a net.Conn that simulates network conditions
// such as latency, loss, duplication, and reordering.
type simulatedConn struct {
	conn    net.Conn
	cfg     *Config
	rand    *rand.Rand
	readBuf []byte
	mu      sync.Mutex

	writeQueue chan []byte
	closeOnce  sync.Once
	closed     chan struct{}
}

// wrapConn wraps an existing net.Conn with simulated network conditions.
func wrapConn(conn net.Conn, cfg *Config) net.Conn {
	sc := &simulatedConn{
		conn:       conn,
		cfg:        cfg,
		rand:       cfg.randSource(),
		writeQueue: make(chan []byte, 100),
		closed:     make(chan struct{}),
	}
	go sc.processWriteQueue()
	return sc
}

// Read reads data from the connection into a buffer, applying network conditions.
func (sc *simulatedConn) Read(b []byte) (int, error) {
	// Simulate loss
	if sc.simulateLoss() {
		// Return an error to simulate a network error
		return 0, io.EOF
	}

	// Read from the underlying connection into a buffer
	buffer := make([]byte, len(b))
	n, err := sc.conn.Read(buffer)
	if n > 0 {
		sc.mu.Lock()

		// Simulate duplication
		if sc.simulateDuplication() {
			sc.readBuf = append(sc.readBuf, buffer[:n]...)
		}

		// Simulate reordering
		if sc.simulateReordering() && len(sc.readBuf) > 0 {
			// Swap the current buffer with the stored buffer
			temp := buffer[:n]
			copy(b, sc.readBuf)
			sc.readBuf = temp
			sc.mu.Unlock()

			// Apply latency
			sc.simulateLatency(n)

			return len(b), nil
		}

		sc.mu.Unlock()

		// Apply latency
		sc.simulateLatency(n)

		// Copy data to the provided slice
		copy(b, buffer[:n])
		return n, err
	}

	return n, err
}

// Write writes data to the connection, applying network conditions.
func (sc *simulatedConn) Write(b []byte) (int, error) {
	// Simulate loss
	if sc.simulateLoss() {
		// Pretend data was sent successfully
		return len(b), nil
	}

	// Simulate duplication
	if sc.simulateDuplication() {
		// Enqueue the data to be sent twice
		dataCopy := append([]byte(nil), b...)
		sc.enqueueWrite(dataCopy)
	}

	// Simulate reordering
	if sc.simulateReordering() {
		// Enqueue the data to be sent later
		dataCopy := append([]byte(nil), b...)
		go func() {
			sc.simulateLatency(len(dataCopy))
			sc.enqueueWrite(dataCopy)
		}()
		return len(b), nil
	}

	// Apply latency
	sc.simulateLatency(len(b))

	// Enqueue the data to be sent
	dataCopy := append([]byte(nil), b...)
	sc.enqueueWrite(dataCopy)

	return len(b), nil
}

// Close closes the connection.
func (sc *simulatedConn) Close() error {
	sc.closeOnce.Do(func() {
		close(sc.closed)
		close(sc.writeQueue)
	})
	return sc.conn.Close()
}

// LocalAddr returns the local network address.
func (sc *simulatedConn) LocalAddr() net.Addr {
	return sc.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (sc *simulatedConn) RemoteAddr() net.Addr {
	return sc.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines.
func (sc *simulatedConn) SetDeadline(t time.Time) error {
	return sc.conn.SetDeadline(t)
}

// SetReadDeadline sets the read deadline.
func (sc *simulatedConn) SetReadDeadline(t time.Time) error {
	return sc.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline.
func (sc *simulatedConn) SetWriteDeadline(t time.Time) error {
	return sc.conn.SetWriteDeadline(t)
}

// simulateLatency applies latency and bandwidth limitations.
func (sc *simulatedConn) simulateLatency(n int) {
	delay := sc.calculateLatency(n)
	if delay > 0 {
		time.Sleep(delay)
	}
}

// calculateLatency calculates the latency based on the network configuration.
func (sc *simulatedConn) calculateLatency(n int) time.Duration {
	latency := sc.cfg.Latency
	if sc.cfg.Jitter > 0 {
		jitter := time.Duration(sc.rand.Int63n(int64(sc.cfg.Jitter)))
		latency += jitter
	}
	if sc.cfg.Bandwidth > 0 && n > 0 {
		transferTime := time.Duration(float64(n) / float64(sc.cfg.Bandwidth) * float64(time.Second))
		latency += transferTime
	}
	return latency
}

// simulateLoss determines if a packet should be dropped based on the loss rate.
func (sc *simulatedConn) simulateLoss() bool {
	return sc.cfg.LossRate > 0 && sc.rand.Float64() < sc.cfg.LossRate
}

// simulateReordering determines if a packet should be reordered based on the reorder rate.
func (sc *simulatedConn) simulateReordering() bool {
	return sc.cfg.ReorderRate > 0 && sc.rand.Float64() < sc.cfg.ReorderRate
}

// simulateDuplication determines if a packet should be duplicated based on the duplicate rate.
func (sc *simulatedConn) simulateDuplication() bool {
	return sc.cfg.DuplicateRate > 0 && sc.rand.Float64() < sc.cfg.DuplicateRate
}

// enqueueWrite enqueues data to be written to the underlying connection.
func (sc *simulatedConn) enqueueWrite(data []byte) {
	select {
	case sc.writeQueue <- data:
	case <-sc.closed:
	}
}

// processWriteQueue processes the write queue, writing data to the underlying connection.
func (sc *simulatedConn) processWriteQueue() {
	for {
		select {
		case data, ok := <-sc.writeQueue:
			if !ok {
				return
			}
			// Write to the underlying connection
			_, err := sc.conn.Write(data)
			if err != nil {
				// Handle error if necessary
			}
		case <-sc.closed:
			return
		}
	}
}
