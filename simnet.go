package simnet

import (
	"math/rand"
	"sync"
	"time"
)

// Config defines the simulated network conditions.
type Config struct {
	mu               sync.Mutex      // Mutex to help ensure thread safety
	rand             *rand.Rand      // Random number generator
	Latency          time.Duration   // Base latency
	Jitter           time.Duration   // Maximum additional latency
	Bandwidth        int64           // Bytes per second (0 means unlimited)
	LossRate         float64         // Packet loss rate (0.0 to 1.0)
	ReorderRate      float64         // Packet reorder rate (0.0 to 1.0)
	DuplicateRate    float64         // Packet duplication rate (0.0 to 1.0)
	PartitionedAddrs map[string]bool // Addresses that are partitioned (unreachable)
	Seed             int64           // Seed for randomness (optional)
}

// Option defines a functional option for configuring network conditions.
type Option func(*Config)

// WithLatency sets the base latency.
func WithLatency(latency time.Duration) Option {
	return func(cfg *Config) {
		cfg.Latency = latency
	}
}

// WithJitter sets the maximum additional latency.
func WithJitter(jitter time.Duration) Option {
	return func(cfg *Config) {
		cfg.Jitter = jitter
	}
}

// WithBandwidth sets the bandwidth limit.
func WithBandwidth(bandwidth int64) Option {
	return func(cfg *Config) {
		cfg.Bandwidth = bandwidth
	}
}

// WithLossRate sets the packet loss rate.
func WithLossRate(lossRate float64) Option {
	return func(cfg *Config) {
		cfg.LossRate = lossRate
	}
}

// WithReorderRate sets the packet reorder rate.
func WithReorderRate(reorderRate float64) Option {
	return func(cfg *Config) {
		cfg.ReorderRate = reorderRate
	}
}

// WithDuplicateRate sets the packet duplication rate.
func WithDuplicateRate(duplicateRate float64) Option {
	return func(cfg *Config) {
		cfg.DuplicateRate = duplicateRate
	}
}

// WithPartitionedAddrs adds partitioned addresses (that are unreachable).
func WithPartitionedAddrs(partitionedAddrs map[string]bool) Option {
	return func(cfg *Config) {
		if cfg.PartitionedAddrs == nil {
			cfg.PartitionedAddrs = make(map[string]bool)
		}
		for addr, val := range partitionedAddrs {
			cfg.PartitionedAddrs[addr] = val
		}
	}
}

// WithSeed sets the seed for randomness.
func WithSeed(seed int64) Option {
	return func(cfg *Config) {
		cfg.Seed = seed
	}
}

// apply applies the options to the config.
func (cfg *Config) apply(opts ...Option) {
	for _, opt := range opts {
		opt(cfg)
	}
}

// NewConfig returns a default NetworkConfig.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		Latency:          0,
		Jitter:           0,
		Bandwidth:        0,
		LossRate:         0.0,
		ReorderRate:      0.0,
		DuplicateRate:    0.0,
		PartitionedAddrs: make(map[string]bool),
	}

	cfg.apply(opts...)

	return cfg
}

// randSource initializes or returns the existing rand.Rand with the given seed.
func (cfg *Config) randSource() *rand.Rand {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	if cfg.rand == nil {
		if cfg.Seed != 0 {
			cfg.rand = rand.New(rand.NewSource(cfg.Seed))
		} else {
			cfg.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
		}
	}
	return cfg.rand
}

// AddPartition adds an address to the partitioned addresses.
func (cfg *Config) AddPartition(address string) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	if cfg.PartitionedAddrs == nil {
		cfg.PartitionedAddrs = make(map[string]bool)
	}
	cfg.PartitionedAddrs[address] = true
}

// RemovePartition removes an address from the partitioned addresses.
func (cfg *Config) RemovePartition(address string) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	delete(cfg.PartitionedAddrs, address)
}
