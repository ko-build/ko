// Package minhash implements the bottom-k sketch for streaming set similarity.
/*

For more information,
    http://research.neustar.biz/2012/07/09/sketch-of-the-day-k-minimum-values/

    MinHashing:
    http://infolab.stanford.edu/~ullman/mmds/ch3.pdf
    https://en.wikipedia.org/wiki/MinHash

    BottomK:
    http://www.math.tau.ac.il/~haimk/papers/p225-cohen.pdf
    http://cohenwang.org/edith/Papers/metrics394-cohen.pdf

    http://www.mpi-inf.mpg.de/~rgemulla/publications/beyer07distinct.pdf

This package works best when provided with a strong 64-bit hash function, such as CityHash, Spooky, Murmur3, or SipHash.

*/
package minhash

import (
	"container/heap"
	"math"
	"sort"
)

type intHeap []uint64

func (h intHeap) Len() int { return len(h) }

// actually Greater, since we want a max-heap
func (h intHeap) Less(i, j int) bool { return h[i] > h[j] }
func (h intHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *intHeap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}

func (h *intHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// BottomK is a bottom-k sketch of a set
type BottomK struct {
	size     int
	h        Hash64
	minimums *intHeap
}

// NewBottomK returns a new BottomK implementation.
func NewBottomK(h Hash64, k int) *BottomK {
	return &BottomK{
		size:     k,
		h:        h,
		minimums: &intHeap{},
	}
}

// Push adds an element to the set.
func (m *BottomK) Push(b []byte) {

	i64 := m.h(b)

	if i64 == 0 {
		return
	}

	if len(*m.minimums) < m.size {
		heap.Push(m.minimums, i64)
		return
	}

	if i64 < (*m.minimums)[0] {
		(*m.minimums)[0] = i64
		heap.Fix(m.minimums, 0)
	}
}

// Merge combines the signatures of the second set, creating the signature of their union.
func (m *BottomK) Merge(m2 *BottomK) {
	for _, v := range *m2.minimums {

		if len(*m.minimums) < m.size {
			heap.Push(m.minimums, v)
			continue
		}

		if v < (*m.minimums)[0] {
			(*m.minimums)[0] = v
			heap.Fix(m.minimums, 0)
		}
	}
}

// Cardinality estimates the cardinality of the set
func (m *BottomK) Cardinality() int {
	return int(float64(len(*m.minimums)-1) / (float64((*m.minimums)[0]) / float64(math.MaxUint64)))

}

// Signature returns a signature for the set.
func (m *BottomK) Signature() []uint64 {
	mins := make(intHeap, len(*m.minimums))
	copy(mins, *m.minimums)
	sort.Sort(mins)
	return mins
}

// Similarity computes an estimate for the similarity between the two sets.
func (m *BottomK) Similarity(m2 *BottomK) float64 {

	if m.size != m2.size {
		panic("minhash minimums size mismatch")
	}

	mins := make(map[uint64]int, len(*m.minimums))

	for _, v := range *m.minimums {
		mins[v]++
	}

	intersect := 0

	for _, v := range *m2.minimums {
		if count, ok := mins[v]; ok && count > 0 {
			intersect++
			mins[v] = count - 1
		}
	}

	maxlength := len(*m.minimums)
	if maxlength < len(*m2.minimums) {
		maxlength = len(*m2.minimums)
	}

	return float64(intersect) / float64(maxlength)
}
