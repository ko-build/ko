package minhashlsh

import (
	"encoding/binary"
	"hash/fnv"
	"math/rand"

	minwise "github.com/dgryski/go-minhash"
)

// The number of byte in a hash value for Minhash
const hashValueSize = 8

// Minhash represents a MinHash object
type Minhash struct {
	mw   *minwise.MinWise
	seed int64
}

// NewMinhash initialize a MinHash object with a seed and the number of
// hash functions.
func NewMinhash(seed int64, numHash int) *Minhash {
	r := rand.New(rand.NewSource(seed))
	b := binary.BigEndian
	b1 := make([]byte, hashValueSize)
	b2 := make([]byte, hashValueSize)
	b.PutUint64(b1, uint64(r.Int63()))
	b.PutUint64(b2, uint64(r.Int63()))
	fnv1 := fnv.New64a()
	fnv2 := fnv.New64a()
	h1 := func(b []byte) uint64 {
		fnv1.Reset()
		fnv1.Write(b1)
		fnv1.Write(b)
		return fnv1.Sum64()
	}
	h2 := func(b []byte) uint64 {
		fnv2.Reset()
		fnv2.Write(b2)
		fnv2.Write(b)
		return fnv2.Sum64()
	}
	return &Minhash{
		mw:   minwise.NewMinWise(h1, h2, numHash),
		seed: int64(seed),
	}
}

// Push a new value to the MinHash object.
// The value should be serialized to byte slice.
func (m *Minhash) Push(b []byte) {
	m.mw.Push(b)
}

// Signature exports the MinHash as a list of hash values.
func (m *Minhash) Signature() []uint64 {
	return m.mw.Signature()
}

// Merge combines the signature of the other Minhash
// with this one, making this one carry the signature of
// the union.
func (m *Minhash) Merge(o *Minhash) {
	if m.seed != o.seed {
		panic("Cannot merge Minhash with different seed")
	}
	m.mw.Merge(o.mw)
}
