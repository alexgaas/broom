#### Implementation of _bloom filter_.

A space-efficient probabilistic data structure for set membership testing with controlled false positive probability.

Simple example of usage:
```go
// Create filter for 10000 elements with 1% false positive rate
bf, err := bloomfilter.New(10000, 0.01)
if err != nil {
    log.Fatal(err)
}

// Add elements
bf.AddString("hello")
bf.AddString("world")

// Check membership
if bf.HasString("hello") {
    fmt.Println("possibly present")
}

if !bf.HasString("unknown") {
    fmt.Println("definitely not present")
}
```

#### Features:

- **Multiple constructors**: `New`, `NewFaster` (power-of-2 bit count), `NewWithHashCount`
- **MurmurHash64** for fast, well-distributed hashing
- **Set operations**: Union, Intersection, SetDifference
- **Serialization**: WriteTo/ReadFrom for persistence
- **Statistics**: BitCount, HashCount, PopCount, EstimateItemCount

#### Pros:

- Production-ready and easily injectable to any Go service
- Thread-safe wrapper available in `bch` package
- Optimal bit count and hash count calculated automatically based on expected elements and error rate

#### Cons:

- May return false positives (by design) - reports element "possibly present" when it's not
- Cannot delete elements from the filter
- Once created, capacity cannot be changed, so initial size of cache must be chosen wisely


