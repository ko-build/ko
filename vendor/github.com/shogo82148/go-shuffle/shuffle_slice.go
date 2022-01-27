//+build go1.8

package shuffle

import (
	"math/rand"
	"reflect"
)

// Slice shuffles the slice.
func Slice(slice interface{}) {
	rv := reflect.ValueOf(slice)
	swap := reflect.Swapper(slice)
	n := rv.Len()
	for i := n - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		swap(i, j)
	}
}

// Slice shuffles the slice.
func (s *Shuffler) Slice(slice Interface) {
	rv := reflect.ValueOf(slice)
	swap := reflect.Swapper(slice)
	n := rv.Len()
	for i := n - 1; i >= 0; i-- {
		j := (*rand.Rand)(s).Intn(i + 1)
		swap(i, j)
	}
}
