package golicenses

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/csv"
	"errors"
	"io"
	"sync"
	"time"
)

//go:embed licenses.csv.gz
var b []byte

var (
	m    map[string]string
	once sync.Once

	// LoadTime is the time it took to load the dataset.
	// It is populated after the first call to Get.
	LoadTime time.Duration

	// NumRecords is the total number of records in the dataset.
	// It is populated after the first call to Get.
	NumRecords int

	ErrNotFound = errors.New("not found")
)

// Get returns the reported license for the package.
//
// The first time Get is called, the dataset is loaded and parsed and stored in
// memory, populating LoadTime and NumRecords. Subsequent calls to Get read
// from memory.
func Get(p string) (string, error) {
	var lerr error
	once.Do(func() {
		start := time.Now()
		m = map[string]string{}
		gr, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			lerr = err
			return
		}
		r := csv.NewReader(gr)
		r.FieldsPerRecord = 2
		for {
			rec, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				lerr = err
				return
			}
			m[rec[0]] = rec[1]
		}

		LoadTime = time.Since(start)
		NumRecords = len(m)
	})
	if lerr != nil {
		return "", lerr
	}

	l, found := m[p]
	if !found {
		return "", ErrNotFound
	}
	return l, nil
}
