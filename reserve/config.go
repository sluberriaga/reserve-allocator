package reserve

import "time"

type AllocatorConfig struct {
	MaxRetryAllocation int
	OvershootFactor    int
	ReserveLifetime    time.Duration
}

type ConcurrencyConfig struct {
	DecayDelay time.Duration
	Decay      uint
	Heat       int
	ConcurrrentThresshold uint64
}

type Config struct {
	Allocator   AllocatorConfig
	Concurrency ConcurrencyConfig
}

func NewConfig() Config {
	return Config{
		Allocator: AllocatorConfig{
			MaxRetryAllocation: 5,
			OvershootFactor:    10,
			ReserveLifetime:    2 * time.Second,
		},
		Concurrency: ConcurrencyConfig{
			DecayDelay: 100 * time.Second,
			Decay:      1,
			Heat:       10,
			ConcurrrentThresshold: 10,
		},
	}
}
