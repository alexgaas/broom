// Package bloomfilter provides a space-efficient probabilistic data structure
// for set membership testing with controlled false positive probability.
package bloomfilter

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"math/bits"
)

// BloomFilter is a space-efficient probabilistic data structure for set membership testing.
// It may return false positives but never false negatives.
type BloomFilter struct {
	hashCount uint64
	bitCount  uint64
	bits      []uint64
}

// New creates a new BloomFilter optimized for the expected number of elements
// and desired false positive rate (error).
//
// Parameters:
//   - expectedElements: expected number of elements to be added
//   - errorRate: desired false positive probability (0 < errorRate < 1)
func New(expectedElements uint64, errorRate float64) (*BloomFilter, error) {
	if errorRate <= 0 || errorRate >= 1 {
		return nil, errors.New("error rate must be between 0 and 1 (exclusive)")
	}
	if expectedElements == 0 {
		expectedElements = 1
	}

	// Calculate optimal bit count: m = -n*ln(p) / (ln(2)^2)
	ln2Squared := math.Ln2 * math.Ln2
	bitCount := uint64(math.Ceil(-float64(expectedElements) * math.Log(errorRate) / ln2Squared))
	bitCount = max(bitCount, 64)

	// Calculate optimal hash count: k = (m/n) * ln(2)
	hashCount := uint64(math.Ceil(float64(bitCount) / float64(expectedElements) * math.Ln2))
	hashCount = max(hashCount, 1)

	return newWithParams(hashCount, bitCount)
}

// NewFaster creates a BloomFilter with bit count rounded to power of 2 for faster operations.
// This uses bitwise AND instead of modulo for index calculation.
func NewFaster(expectedElements uint64, errorRate float64) (*BloomFilter, error) {
	if errorRate <= 0 || errorRate >= 1 {
		return nil, errors.New("error rate must be between 0 and 1 (exclusive)")
	}
	if expectedElements == 0 {
		expectedElements = 1
	}

	// Calculate bit count and round to power of 2
	ln2Squared := math.Ln2 * math.Ln2
	fractionalBitCount := -float64(expectedElements) * math.Log(errorRate) / ln2Squared
	fractionalBitCount = math.Max(64, fractionalBitCount)

	// Round up to next power of 2
	bitCount := uint64(1) << uint64(math.Ceil(math.Log2(fractionalBitCount)))

	// Calculate hash count
	hashCount := uint64(math.Ceil(math.Max(1, float64(bitCount)/float64(expectedElements)*math.Ln2)))

	return newWithParams(hashCount, bitCount)
}

// NewWithHashCount creates a BloomFilter with explicit hash count.
// More hash functions = faster but more memory.
func NewWithHashCount(expectedElements, hashCount uint64, errorRate float64) (*BloomFilter, error) {
	if errorRate <= 0 || errorRate >= 1 {
		return nil, errors.New("error rate must be between 0 and 1 (exclusive)")
	}
	if expectedElements == 0 {
		expectedElements = 1
	}
	if hashCount == 0 {
		hashCount = 1
	}

	// Calculate bit count for given hash count: m = -1 / ln(1 - p^(1/k)) * k * n
	bitCount := uint64(math.Ceil(-1.0 / math.Log(1.0-math.Pow(errorRate, 1.0/float64(hashCount))) * float64(hashCount) * float64(expectedElements)))
	bitCount = max(bitCount, 64)

	return newWithParams(hashCount, bitCount)
}

func newWithParams(hashCount, bitCount uint64) (*BloomFilter, error) {
	bf := &BloomFilter{
		hashCount: hashCount,
		bitCount:  bitCount,
		bits:      make([]uint64, (bitCount+63)/64),
	}
	return bf, nil
}

// Random seeds for hash generation
var randomSeeds = []uint64{
	0x4b7db4c869874dd1,
	0x43e9b39115fd33ba,
	0x180be656098797e4,
	0xe21f17e9d2d0bae7,
	0xeaa42039facc7152,
	0x0d3666daa04ff2fd,
	0xcafb2c5b513bc4f0,
	0xdbb86c8c0293d7be,
	0xec978c08a6a50237,
	0xa3601812c4207d5d,
	0x6ad826554038feae,
	0xebe6d4db55c4f77a,
	0xe976cceb2abbc306,
	0xeac2796bf5c2907b,
	0x0673b5ce5e1e40fd,
	0x004e6cc070604495,
}

// murmurHash64 implements MurmurHash64A
func murmurHash64(data []byte, seed uint64) uint64 {
	const m uint64 = 0xc6a4a7935bd1e995
	const r = 47

	h := seed ^ (uint64(len(data)) * m)

	// Process 8 bytes at a time
	nblocks := len(data) / 8
	for i := 0; i < nblocks; i++ {
		k := binary.LittleEndian.Uint64(data[i*8:])
		k *= m
		k ^= k >> r
		k *= m
		h ^= k
		h *= m
	}

	// Process remaining bytes
	tail := data[nblocks*8:]
	switch len(tail) {
	case 7:
		h ^= uint64(tail[6]) << 48
		fallthrough
	case 6:
		h ^= uint64(tail[5]) << 40
		fallthrough
	case 5:
		h ^= uint64(tail[4]) << 32
		fallthrough
	case 4:
		h ^= uint64(tail[3]) << 24
		fallthrough
	case 3:
		h ^= uint64(tail[2]) << 16
		fallthrough
	case 2:
		h ^= uint64(tail[1]) << 8
		fallthrough
	case 1:
		h ^= uint64(tail[0])
		h *= m
	}

	h ^= h >> r
	h *= m
	h ^= h >> r

	return h
}

// getHashBits generates hash values for the bloom filter
func (bf *BloomFilter) getHashBits(data []byte, index uint64) uint64 {
	seed := randomSeeds[index%uint64(len(randomSeeds))] + index
	return murmurHash64(data, seed) % bf.bitCount
}

// setBit sets a bit in the bit array
func (bf *BloomFilter) setBit(index uint64) {
	bf.bits[index>>6] |= 1 << (index & 63)
}

// getBit returns true if a bit is set
func (bf *BloomFilter) getBit(index uint64) bool {
	return bf.bits[index>>6]&(1<<(index&63)) != 0
}

// Add adds an element to the bloom filter.
func (bf *BloomFilter) Add(data []byte) {
	for i := uint64(0); i < bf.hashCount; i++ {
		bitIndex := bf.getHashBits(data, i)
		bf.setBit(bitIndex)
	}
}

// AddString adds a string element to the bloom filter.
func (bf *BloomFilter) AddString(s string) {
	bf.Add([]byte(s))
}

// Has checks if an element might be in the set.
// Returns true if the element is possibly in the set (may be false positive).
// Returns false if the element is definitely not in the set.
func (bf *BloomFilter) Has(data []byte) bool {
	for i := uint64(0); i < bf.hashCount; i++ {
		bitIndex := bf.getHashBits(data, i)
		if !bf.getBit(bitIndex) {
			return false
		}
	}
	return true
}

// HasString checks if a string element might be in the set.
func (bf *BloomFilter) HasString(s string) bool {
	return bf.Has([]byte(s))
}

// Clear resets all bits to zero.
func (bf *BloomFilter) Clear() {
	for i := range bf.bits {
		bf.bits[i] = 0
	}
}

// IsEmpty returns true if no bits are set.
func (bf *BloomFilter) IsEmpty() bool {
	for _, v := range bf.bits {
		if v != 0 {
			return false
		}
	}
	return true
}

// BitCount returns the number of bits in the filter.
func (bf *BloomFilter) BitCount() uint64 {
	return bf.bitCount
}

// HashCount returns the number of hash functions used.
func (bf *BloomFilter) HashCount() uint64 {
	return bf.hashCount
}

// PopCount returns the number of bits set to 1.
func (bf *BloomFilter) PopCount() uint64 {
	var count uint64
	for _, v := range bf.bits {
		count += uint64(bits.OnesCount64(v))
	}
	return count
}

// EstimateItemCount estimates the number of items in the filter.
// Based on the formula: n* = -m/k * ln(1 - X/m) where X is the number of set bits.
func (bf *BloomFilter) EstimateItemCount() float64 {
	setBits := bf.PopCount()
	if setBits == bf.bitCount {
		// All bits set, use approximation
		return -float64(bf.bitCount) * math.Log(0.001) / float64(bf.hashCount)
	}
	return -float64(bf.bitCount) * math.Log(1.0-float64(setBits)/float64(bf.bitCount)) / float64(bf.hashCount)
}

// Union performs a union operation with another BloomFilter.
// Both filters must have the same parameters.
func (bf *BloomFilter) Union(other *BloomFilter) error {
	if err := bf.checkCompatible(other); err != nil {
		return err
	}
	for i := range bf.bits {
		bf.bits[i] |= other.bits[i]
	}
	return nil
}

// Intersection performs an intersection operation with another BloomFilter.
// Both filters must have the same parameters.
func (bf *BloomFilter) Intersection(other *BloomFilter) error {
	if err := bf.checkCompatible(other); err != nil {
		return err
	}
	for i := range bf.bits {
		bf.bits[i] &= other.bits[i]
	}
	return nil
}

// SetDifference performs a set difference operation (this - other).
// Both filters must have the same parameters.
func (bf *BloomFilter) SetDifference(other *BloomFilter) error {
	if err := bf.checkCompatible(other); err != nil {
		return err
	}
	for i := range bf.bits {
		bf.bits[i] &^= other.bits[i]
	}
	return nil
}

func (bf *BloomFilter) checkCompatible(other *BloomFilter) error {
	if bf.hashCount != other.hashCount {
		return errors.New("hash count mismatch")
	}
	if bf.bitCount != other.bitCount {
		return errors.New("bit count mismatch")
	}
	if len(bf.bits) != len(other.bits) {
		return errors.New("bits array size mismatch")
	}
	return nil
}

// Clone creates a deep copy of the BloomFilter.
func (bf *BloomFilter) Clone() *BloomFilter {
	newBits := make([]uint64, len(bf.bits))
	copy(newBits, bf.bits)
	return &BloomFilter{
		hashCount: bf.hashCount,
		bitCount:  bf.bitCount,
		bits:      newBits,
	}
}

// WriteTo writes the bloom filter to a writer.
// Format: hashCount (8 bytes) | bitCount (8 bytes) | bits (variable)
func (bf *BloomFilter) WriteTo(w io.Writer) (int64, error) {
	var written int64

	// Write header
	header := make([]byte, 16)
	binary.LittleEndian.PutUint64(header[0:8], bf.hashCount)
	binary.LittleEndian.PutUint64(header[8:16], bf.bitCount)
	n, err := w.Write(header)
	written += int64(n)
	if err != nil {
		return written, err
	}

	// Write bits
	for _, v := range bf.bits {
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], v)
		n, err := w.Write(buf[:])
		written += int64(n)
		if err != nil {
			return written, err
		}
	}

	return written, nil
}

// ReadFrom reads a bloom filter from a reader.
func (bf *BloomFilter) ReadFrom(r io.Reader) (int64, error) {
	var read int64

	// Read header
	header := make([]byte, 16)
	n, err := io.ReadFull(r, header)
	read += int64(n)
	if err != nil {
		return read, err
	}

	bf.hashCount = binary.LittleEndian.Uint64(header[0:8])
	bf.bitCount = binary.LittleEndian.Uint64(header[8:16])

	// Allocate and read bits
	bf.bits = make([]uint64, (bf.bitCount+63)/64)
	for i := range bf.bits {
		var buf [8]byte
		n, err := io.ReadFull(r, buf[:])
		read += int64(n)
		if err != nil {
			return read, err
		}
		bf.bits[i] = binary.LittleEndian.Uint64(buf[:])
	}

	return read, nil
}

// GetOptimalBitCount returns the optimal bit count for given parameters.
func GetOptimalBitCount(expectedElements uint64, errorRate float64) uint64 {
	if errorRate <= 0 || errorRate >= 1 || expectedElements == 0 {
		return 64
	}
	ln2Squared := math.Ln2 * math.Ln2
	return uint64(math.Ceil(-float64(expectedElements) * math.Log(errorRate) / ln2Squared))
}

// GetOptimalHashCount returns the optimal hash count for given error rate.
func GetOptimalHashCount(errorRate float64) uint64 {
	if errorRate <= 0 || errorRate >= 1 {
		return 1
	}
	return uint64(math.Ceil(-math.Log2(errorRate)))
}
