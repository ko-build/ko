package minhash

import "math"

// MinWise is a collection of minimum hashes for a set
type MinWise struct {
	minimums []uint64
	h1       Hash64
	h2       Hash64
}

type Hash64 func([]byte) uint64

// NewMinWise returns a new MinWise Hashsing implementation
func NewMinWise(h1, h2 Hash64, size int) *MinWise {

	minimums := make([]uint64, size)
	for i := range minimums {
		minimums[i] = math.MaxUint64
	}

	return &MinWise{
		h1:       h1,
		h2:       h2,
		minimums: minimums,
	}
}

// Push adds an element to the set.
func (m *MinWise) Push(b []byte) {

	v1 := m.h1(b)
	v2 := m.h2(b)

	for i, v := range m.minimums {
		hv := v1 + uint64(i)*v2
		if hv < v {
			m.minimums[i] = hv
		}
	}
}

// Merge combines the signatures of the second set, creating the signature of their union.
func (m *MinWise) Merge(m2 *MinWise) {

	for i, v := range m2.minimums {

		if v < m.minimums[i] {
			m.minimums[i] = v
		}
	}
}

// Cardinality estimates the cardinality of the set
func (m *MinWise) Cardinality() int {

	// http://www.cohenwang.com/edith/Papers/tcest.pdf

	sum := 0.0

	for _, v := range m.minimums {
		sum += -math.Log(float64(math.MaxUint64-v) / float64(math.MaxUint64))
	}

	return int(float64(len(m.minimums)-1) / sum)
}

// Signature returns a signature for the set.
func (m *MinWise) Signature() []uint64 {
	return m.minimums
}

// Similarity computes an estimate for the similarity between the two sets.
func (m *MinWise) Similarity(m2 *MinWise) float64 {

	if len(m.minimums) != len(m2.minimums) {
		panic("minhash minimums size mismatch")
	}

	intersect := 0

	for i := range m.minimums {
		if m.minimums[i] == m2.minimums[i] {
			intersect++
		}
	}

	return float64(intersect) / float64(len(m.minimums))
}

// SignatureBbit returns a b-bit reduction of the signature.  This will result in unused bits at the high-end of the words if b does not divide 64 evenly.
func (m *MinWise) SignatureBbit(b uint) []uint64 {

	var sig []uint64 // full signature
	var w uint64     // current word
	bits := uint(64) // bits free in current word

	mask := uint64(1<<b) - 1

	for _, v := range m.minimums {
		if bits >= b {
			w <<= b
			w |= v & mask
			bits -= b
		} else {
			sig = append(sig, w)
			w = 0
			bits = 64
		}
	}

	if bits != 64 {
		sig = append(sig, w)
	}

	return sig
}

// SimilarityBbit computes an estimate for the similarity between two b-bit signatures
func SimilarityBbit(sig1, sig2 []uint64, b uint) float64 {

	if len(sig1) != len(sig2) {
		panic("signature size mismatch")
	}

	intersect := 0
	count := 0

	mask := uint64(1<<b) - 1

	for i := range sig1 {
		w1 := sig1[i]
		w2 := sig2[i]

		bits := uint(64)

		for bits >= b {
			v1 := (w1 & mask)
			v2 := (w2 & mask)

			count++
			if v1 == v2 {
				intersect++
			}

			bits -= b
			w1 >>= b
			w2 >>= b
		}
	}

	return float64(intersect) / float64(count)
}
