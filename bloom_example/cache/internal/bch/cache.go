package bch

import (
	"sync"

	"github.com/alexgaas/bloomfilter"
)

// BloomCache wraps a bloom filter with thread-safe operations
type BloomCache struct {
	filter *bloomfilter.BloomFilter
	mu     sync.RWMutex

	expectedElements uint64
	errorRate        float64
}

// Config holds bloom cache configuration
type Config struct {
	ExpectedElements uint64  `json:"expected_elements" mapstructure:"expected_elements"`
	ErrorRate        float64 `json:"error_rate" mapstructure:"error_rate"`
}

// DefaultConfig returns default bloom cache configuration
func DefaultConfig() Config {
	return Config{
		ExpectedElements: 1000000,
		ErrorRate:        0.01,
	}
}

// New creates a new BloomCache with given configuration
func New(cfg Config) (*BloomCache, error) {
	filter, err := bloomfilter.New(cfg.ExpectedElements, cfg.ErrorRate)
	if err != nil {
		return nil, err
	}

	return &BloomCache{
		filter:           filter,
		expectedElements: cfg.ExpectedElements,
		errorRate:        cfg.ErrorRate,
	}, nil
}

// Add adds a key to the bloom cache
func (bc *BloomCache) Add(key string) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.filter.AddString(key)
}

// AddBytes adds a byte slice key to the bloom cache
func (bc *BloomCache) AddBytes(key []byte) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.filter.Add(key)
}

// Has checks if a key might exist in the bloom cache
// Returns true if possibly present (may be false positive)
// Returns false if definitely not present
func (bc *BloomCache) Has(key string) bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.filter.HasString(key)
}

// HasBytes checks if a byte slice key might exist
func (bc *BloomCache) HasBytes(key []byte) bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.filter.Has(key)
}

// Clear resets the bloom filter
func (bc *BloomCache) Clear() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.filter.Clear()
}

// Stats returns current bloom cache statistics
type Stats struct {
	ExpectedElements   uint64  `json:"expected_elements"`
	ErrorRate          float64 `json:"error_rate"`
	BitCount           uint64  `json:"bit_count"`
	HashCount          uint64  `json:"hash_count"`
	BitsSet            uint64  `json:"bits_set"`
	EstimatedItemCount float64 `json:"estimated_item_count"`
	IsEmpty            bool    `json:"is_empty"`
}

// Stats returns statistics about the bloom cache
func (bc *BloomCache) Stats() Stats {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return Stats{
		ExpectedElements:   bc.expectedElements,
		ErrorRate:          bc.errorRate,
		BitCount:           bc.filter.BitCount(),
		HashCount:          bc.filter.HashCount(),
		BitsSet:            bc.filter.PopCount(),
		EstimatedItemCount: bc.filter.EstimateItemCount(),
		IsEmpty:            bc.filter.IsEmpty(),
	}
}
