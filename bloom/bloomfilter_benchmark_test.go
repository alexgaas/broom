package bloomfilter_test

import (
	"fmt"
	"testing"

	. "github.com/alexgaas/bloomfilter"
)

func BenchmarkAdd(b *testing.B) {
	bf, _ := New(uint64(b.N), 0.01)
	data := []byte("benchmark_test_data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add(data)
	}
}

func BenchmarkHas(b *testing.B) {
	bf, _ := New(10000, 0.01)
	data := []byte("benchmark_test_data")
	bf.Add(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Has(data)
	}
}

func BenchmarkAddString(b *testing.B) {
	bf, _ := New(uint64(b.N), 0.01)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.AddString("benchmark_test_string")
	}
}

func BenchmarkHasString(b *testing.B) {
	bf, _ := New(10000, 0.01)
	bf.AddString("benchmark_test_string")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.HasString("benchmark_test_string")
	}
}

func BenchmarkAddFaster(b *testing.B) {
	bf, _ := NewFaster(uint64(b.N), 0.01)
	data := []byte("benchmark_test_data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add(data)
	}
}

func BenchmarkHasFaster(b *testing.B) {
	bf, _ := NewFaster(10000, 0.01)
	data := []byte("benchmark_test_data")
	bf.Add(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Has(data)
	}
}

func BenchmarkUnion(b *testing.B) {
	bf1, _ := New(10000, 0.01)
	bf2, _ := New(10000, 0.01)

	for i := 0; i < 5000; i++ {
		bf1.AddString(fmt.Sprintf("item_%d", i))
		bf2.AddString(fmt.Sprintf("other_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clone := bf1.Clone()
		clone.Union(bf2)
	}
}

func BenchmarkEstimateItemCount(b *testing.B) {
	bf, _ := New(10000, 0.01)
	for i := 0; i < 5000; i++ {
		bf.AddString(fmt.Sprintf("item_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.EstimateItemCount()
	}
}

func BenchmarkPopCount(b *testing.B) {
	bf, _ := New(100000, 0.001)
	for i := 0; i < 50000; i++ {
		bf.AddString(fmt.Sprintf("item_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.PopCount()
	}
}

// Benchmark with different sizes
func BenchmarkAddSize100(b *testing.B)    { benchmarkAddSize(b, 100) }
func BenchmarkAddSize1000(b *testing.B)   { benchmarkAddSize(b, 1000) }
func BenchmarkAddSize10000(b *testing.B)  { benchmarkAddSize(b, 10000) }
func BenchmarkAddSize100000(b *testing.B) { benchmarkAddSize(b, 100000) }

func benchmarkAddSize(b *testing.B, size uint64) {
	bf, _ := New(size, 0.01)
	data := []byte("benchmark_test_data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add(data)
	}
}

// Benchmark with different error rates
func BenchmarkAddError001(b *testing.B)   { benchmarkAddError(b, 0.01) }
func BenchmarkAddError0001(b *testing.B)  { benchmarkAddError(b, 0.001) }
func BenchmarkAddError00001(b *testing.B) { benchmarkAddError(b, 0.0001) }

func benchmarkAddError(b *testing.B, errorRate float64) {
	bf, _ := New(10000, errorRate)
	data := []byte("benchmark_test_data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add(data)
	}
}
