package bloomfilter_test

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	. "github.com/alexgaas/bloomfilter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBloomFilter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BloomFilter Suite")
}

var _ = Describe("BloomFilter", func() {

	Describe("Constructor", func() {
		Context("with valid parameters", func() {
			It("should create a new BloomFilter", func() {
				bf, err := New(1000, 0.01)
				Expect(err).NotTo(HaveOccurred())
				Expect(bf).NotTo(BeNil())
				Expect(bf.BitCount()).To(BeNumerically(">", 0))
				Expect(bf.HashCount()).To(BeNumerically(">", 0))
			})

			It("should handle zero elements by defaulting to 1", func() {
				bf, err := New(0, 0.01)
				Expect(err).NotTo(HaveOccurred())
				Expect(bf).NotTo(BeNil())
			})
		})

		Context("with invalid parameters", func() {
			It("should reject error rate of 0", func() {
				_, err := New(1000, 0)
				Expect(err).To(HaveOccurred())
			})

			It("should reject error rate of 1", func() {
				_, err := New(1000, 1)
				Expect(err).To(HaveOccurred())
			})

			It("should reject negative error rate", func() {
				_, err := New(1000, -0.1)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("NewFaster", func() {
		It("should create a filter with power-of-2 bit count", func() {
			bf, err := NewFaster(1000, 0.01)
			Expect(err).NotTo(HaveOccurred())

			// Check that BitCount is a power of 2
			bitCount := bf.BitCount()
			Expect(bitCount & (bitCount - 1)).To(Equal(uint64(0)))
		})

		It("should work correctly for add and has operations", func() {
			bf, _ := NewFaster(1000, 0.01)
			bf.AddString("test")
			Expect(bf.HasString("test")).To(BeTrue())
		})
	})

	Describe("NewWithHashCount", func() {
		It("should create a filter with specified hash count", func() {
			const hashCount uint64 = 3
			bf, err := NewWithHashCount(1000, hashCount, 0.01)
			Expect(err).NotTo(HaveOccurred())
			Expect(bf.HashCount()).To(Equal(hashCount))
		})

		It("should work correctly for add and has operations", func() {
			bf, _ := NewWithHashCount(1000, 3, 0.01)
			bf.AddString("test")
			Expect(bf.HasString("test")).To(BeTrue())
		})
	})

	Describe("Add and Has", func() {
		var bf *BloomFilter

		BeforeEach(func() {
			bf, _ = New(1000, 0.01)
		})

		Context("with string items", func() {
			It("should find added items", func() {
				items := []string{"hello", "world", "bloom", "filter"}
				for _, item := range items {
					bf.AddString(item)
				}

				for _, item := range items {
					Expect(bf.HasString(item)).To(BeTrue(), "Expected to find %q", item)
				}
			})

			It("should handle empty string", func() {
				bf.AddString("")
				Expect(bf.HasString("")).To(BeTrue())
			})

			It("should handle special characters", func() {
				special := "Special!@#$%^&*()chars"
				bf.AddString(special)
				Expect(bf.HasString(special)).To(BeTrue())
			})

			It("should handle unicode strings", func() {
				unicode := "Unicode: こんにちは 🎉"
				bf.AddString(unicode)
				Expect(bf.HasString(unicode)).To(BeTrue())
			})
		})

		Context("with byte slices", func() {
			It("should find added byte slices", func() {
				data := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
				bf.Add(data)
				Expect(bf.Has(data)).To(BeTrue())
			})
		})
	})

	Describe("False Negative Guarantee", func() {
		It("should never have false negatives", func() {
			bf, _ := New(1000, 0.01)

			items := make([]string, 500)
			for i := 0; i < 500; i++ {
				items[i] = fmt.Sprintf("item_%d", i)
			}

			for _, item := range items {
				bf.AddString(item)
			}

			for _, item := range items {
				Expect(bf.HasString(item)).To(BeTrue(), "False negative for %q", item)
			}
		})
	})

	Describe("False Positive Rate", func() {
		It("should maintain acceptable false positive rate", func() {
			const n = 10000
			const targetFPR = 0.01

			bf, _ := New(n, targetFPR)

			for i := 0; i < n; i++ {
				bf.AddString(fmt.Sprintf("added_%d", i))
			}

			var falsePositives int
			const testCount = 100000

			for i := 0; i < testCount; i++ {
				if bf.HasString(fmt.Sprintf("not_added_%d", i)) {
					falsePositives++
				}
			}

			actualFPR := float64(falsePositives) / testCount

			// Allow 3x margin for statistical variance
			Expect(actualFPR).To(BeNumerically("<", targetFPR*3))

			GinkgoWriter.Printf("Actual FPR: %f (target: %f)\n", actualFPR, targetFPR)
		})
	})

	Describe("Clear", func() {
		It("should reset all bits to zero", func() {
			bf, _ := New(100, 0.01)

			bf.AddString("test1")
			bf.AddString("test2")

			Expect(bf.IsEmpty()).To(BeFalse())
			Expect(bf.PopCount()).To(BeNumerically(">", 0))

			bf.Clear()

			Expect(bf.IsEmpty()).To(BeTrue())
			Expect(bf.PopCount()).To(Equal(uint64(0)))
		})
	})

	Describe("IsEmpty", func() {
		It("should return true for new filter", func() {
			bf, _ := New(100, 0.01)
			Expect(bf.IsEmpty()).To(BeTrue())
		})

		It("should return false after adding items", func() {
			bf, _ := New(100, 0.01)
			bf.AddString("test")
			Expect(bf.IsEmpty()).To(BeFalse())
		})
	})

	Describe("Optimal Parameters", func() {
		It("should calculate optimal bit count", func() {
			bf, _ := New(1000, 0.01)

			// For n=1000, p=0.01: optimal m ≈ 9585 bits
			Expect(bf.BitCount()).To(BeNumerically(">=", 9000))
			Expect(bf.BitCount()).To(BeNumerically("<=", 10000))
		})

		It("should calculate optimal hash count", func() {
			bf, _ := New(1000, 0.01)

			// For n=1000, p=0.01: optimal k ≈ 7
			Expect(bf.HashCount()).To(BeNumerically(">=", 5))
			Expect(bf.HashCount()).To(BeNumerically("<=", 10))
		})
	})

	Describe("Large Number of Items", func() {
		It("should handle 50000 items without false negatives", func() {
			bf, _ := New(100000, 0.001)

			for i := 0; i < 50000; i++ {
				bf.AddString(fmt.Sprintf("large_test_%d", i))
			}

			for i := 0; i < 50000; i++ {
				Expect(bf.HasString(fmt.Sprintf("large_test_%d", i))).To(BeTrue())
			}
		})
	})

	Describe("Set Operations", func() {
		var bf1, bf2 *BloomFilter

		BeforeEach(func() {
			bf1, _ = New(100, 0.01)
			bf2, _ = New(100, 0.01)
		})

		Describe("Union", func() {
			It("should combine elements from both filters", func() {
				bf1.AddString("a")
				bf1.AddString("b")
				bf2.AddString("c")
				bf2.AddString("d")

				err := bf1.Union(bf2)
				Expect(err).NotTo(HaveOccurred())

				Expect(bf1.HasString("a")).To(BeTrue())
				Expect(bf1.HasString("b")).To(BeTrue())
				Expect(bf1.HasString("c")).To(BeTrue())
				Expect(bf1.HasString("d")).To(BeTrue())
			})
		})

		Describe("Intersection", func() {
			It("should keep only common elements", func() {
				bf1.AddString("a")
				bf1.AddString("common")
				bf2.AddString("b")
				bf2.AddString("common")

				err := bf1.Intersection(bf2)
				Expect(err).NotTo(HaveOccurred())

				Expect(bf1.HasString("common")).To(BeTrue())
			})
		})

		Describe("SetDifference", func() {
			It("should remove elements in second filter", func() {
				bf1.AddString("a")
				bf1.AddString("b")
				bf2.AddString("b")

				originalPopCount := bf1.PopCount()

				err := bf1.SetDifference(bf2)
				Expect(err).NotTo(HaveOccurred())

				Expect(bf1.PopCount()).To(BeNumerically("<", originalPopCount))
			})
		})

		Describe("Incompatible Filters", func() {
			It("should return error for filters with different sizes", func() {
				bf3, _ := New(1000, 0.01) // Different size

				err := bf1.Union(bf3)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Clone", func() {
		It("should create an independent copy", func() {
			bf1, _ := New(100, 0.01)
			bf1.AddString("test")

			bf2 := bf1.Clone()

			Expect(bf2.HasString("test")).To(BeTrue())

			bf2.Clear()
			Expect(bf1.HasString("test")).To(BeTrue(), "Original should not be affected")
		})
	})

	Describe("EstimateItemCount", func() {
		It("should estimate within 10% of actual count", func() {
			bf, _ := New(10000, 0.01)

			const addCount = 5000
			for i := 0; i < addCount; i++ {
				bf.AddString(fmt.Sprintf("estimate_%d", i))
			}

			estimate := bf.EstimateItemCount()
			errorPercent := math.Abs(estimate-addCount) / addCount * 100

			Expect(errorPercent).To(BeNumerically("<", 10))

			GinkgoWriter.Printf("Added: %d, Estimated: %.0f (%.1f%% error)\n", addCount, estimate, errorPercent)
		})
	})

	Describe("Serialization", func() {
		It("should serialize and deserialize correctly", func() {
			bf1, _ := New(1000, 0.01)

			items := []string{"serialize", "test", "bloom", "filter"}
			for _, item := range items {
				bf1.AddString(item)
			}

			var buf bytes.Buffer
			_, err := bf1.WriteTo(&buf)
			Expect(err).NotTo(HaveOccurred())

			bf2 := &BloomFilter{}
			_, err = bf2.ReadFrom(&buf)
			Expect(err).NotTo(HaveOccurred())

			Expect(bf2.HashCount()).To(Equal(bf1.HashCount()))
			Expect(bf2.BitCount()).To(Equal(bf1.BitCount()))

			for _, item := range items {
				Expect(bf2.HasString(item)).To(BeTrue(), "Missing %q after deserialization", item)
			}
		})
	})

	Describe("Helper Functions", func() {
		Describe("GetOptimalBitCount", func() {
			It("should return expected value for known parameters", func() {
				bitCount := GetOptimalBitCount(1000, 0.01)
				Expect(bitCount).To(BeNumerically(">=", 9000))
				Expect(bitCount).To(BeNumerically("<=", 10000))
			})
		})

		Describe("GetOptimalHashCount", func() {
			It("should return expected value for known parameters", func() {
				hashCount := GetOptimalHashCount(0.01)
				Expect(hashCount).To(BeNumerically(">=", 6))
				Expect(hashCount).To(BeNumerically("<=", 8))
			})
		})
	})
})
