package wmh

import (
	"encoding/binary"
	"errors"
	"log"
	"math"
	"reflect"
	"unsafe"

	"github.com/go-enry/go-license-detector/v4/licensedb/internal/fastlog"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

const maxUint16 = 65535

// WeightedMinHasher calculates Weighted MinHash-es.
// https://ekzhu.github.io/datasketch/weightedminhash.html
type WeightedMinHasher struct {
	// Size of each hash element in bits. Supported values are 16, 32 and 64.
	Bitness int

	dim        int
	sampleSize int
	rs         [][]float32
	lnCs       [][]float32
	betas      [][]uint16 // attempt to save some memory - [0, 1] is scaled to maxUint16
}

// NewWeightedMinHasher initializes a new instance of WeightedMinHasher.
// `dim` is the bag size.
// `sampleSize` is the hash length.
// `seed` is the random generator seed, as Weighted MinHash is probabilistic.
func NewWeightedMinHasher(dim int, sampleSize int, seed int64) *WeightedMinHasher {
	randSrc := rand.New(rand.NewSource(uint64(seed)))
	gammaGen := distuv.Gamma{Alpha: 2, Beta: 1, Src: randSrc}
	hasher := &WeightedMinHasher{Bitness: 64, dim: dim, sampleSize: sampleSize}
	hasher.rs = make([][]float32, sampleSize)
	for y := 0; y < sampleSize; y++ {
		arr := make([]float32, dim)
		hasher.rs[y] = arr
		for x := 0; x < dim; x++ {
			arr[x] = float32(gammaGen.Rand())
		}
	}
	hasher.lnCs = make([][]float32, sampleSize)
	for y := 0; y < sampleSize; y++ {
		arr := make([]float32, dim)
		hasher.lnCs[y] = arr
		for x := 0; x < dim; x++ {
			arr[x] = fastlog.Log(float32(gammaGen.Rand()))
		}
	}
	uniformGen := distuv.Uniform{Min: 0, Max: 1, Src: randSrc}
	hasher.betas = make([][]uint16, sampleSize)
	for y := 0; y < sampleSize; y++ {
		arr := make([]uint16, dim)
		hasher.betas[y] = arr
		for x := 0; x < dim; x++ {
			arr[x] = uint16(uniformGen.Rand() * maxUint16)
		}
	}
	return hasher
}

// MarshalBinary serializes the WeightedMinHasher.
func (wmh *WeightedMinHasher) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 9+wmh.sampleSize*wmh.dim*(4*2+2))
	data[0] = byte(wmh.Bitness)
	binary.LittleEndian.PutUint32(data[1:5], uint32(wmh.dim))
	binary.LittleEndian.PutUint32(data[5:9], uint32(wmh.sampleSize))
	offset := 9
	writeFloat32Slice := func(arr []float32) {
		header := *(*reflect.SliceHeader)(unsafe.Pointer(&arr))
		header.Len *= 4
		header.Cap *= 4
		buffer := *(*[]byte)(unsafe.Pointer(&header))
		copy(data[offset:], buffer)
		offset += len(buffer)
	}
	for _, arr := range wmh.rs {
		writeFloat32Slice(arr)
	}
	for _, arr := range wmh.lnCs {
		writeFloat32Slice(arr)
	}
	for _, arr := range wmh.betas {
		header := *(*reflect.SliceHeader)(unsafe.Pointer(&arr))
		header.Len *= 2
		header.Cap *= 2
		buffer := *(*[]byte)(unsafe.Pointer(&header))
		copy(data[offset:], buffer)
		offset += len(buffer)
	}
	return data, nil
}

// UnmarshalBinary reads a WeightedMinHasher previously serialized with MarshalBinary().
func (wmh *WeightedMinHasher) UnmarshalBinary(data []byte) error {
	if len(data) < 9 {
		return errors.New("invalid binary format: no header")
	}
	wmh.Bitness = int(data[0])
	wmh.dim = int(binary.LittleEndian.Uint32(data[1:5]))
	wmh.sampleSize = int(binary.LittleEndian.Uint32(data[5:9]))
	if len(data)-9 != wmh.sampleSize*wmh.dim*(4*2+2) {
		return errors.New("invalid binary format: body size mismatch")
	}
	wmh.rs = make([][]float32, wmh.sampleSize)
	wmh.lnCs = make([][]float32, wmh.sampleSize)
	wmh.betas = make([][]uint16, wmh.sampleSize)
	readFloat32Slice := func(dest []float32, src []byte) {
		header := *(*reflect.SliceHeader)(unsafe.Pointer(&src))
		header.Len /= 4
		header.Cap /= 4
		buffer := *(*[]float32)(unsafe.Pointer(&header))
		copy(dest, buffer)
	}
	offset := 9
	for i := range wmh.rs {
		wmh.rs[i] = make([]float32, wmh.dim)
		nextOffset := offset + wmh.dim*4
		readFloat32Slice(wmh.rs[i], data[offset:nextOffset])
		offset = nextOffset
	}
	for i := range wmh.lnCs {
		wmh.lnCs[i] = make([]float32, wmh.dim)
		nextOffset := offset + wmh.dim*4
		readFloat32Slice(wmh.lnCs[i], data[offset:nextOffset])
		offset = nextOffset
	}
	for i := range wmh.betas {
		wmh.betas[i] = make([]uint16, wmh.dim)
		nextOffset := offset + wmh.dim*2
		slice := data[offset:nextOffset]
		header := *(*reflect.SliceHeader)(unsafe.Pointer(&slice))
		header.Len /= 2
		header.Cap /= 2
		buffer := *(*[]uint16)(unsafe.Pointer(&header))
		copy(wmh.betas[i], buffer)
		offset = nextOffset
	}
	return nil
}

// Hash calculates the Weighted MinHash from the weighted bag of features.
// Each feature has an index and a value.
func (wmh *WeightedMinHasher) Hash(values []float32, indices []int) []uint64 {
	if len(values) != len(indices) {
		log.Panicf("len(values)=%d is not equal to len(indices)=%d", len(values), len(indices))
	}
	for i, v := range values {
		if v < 0 {
			log.Panicf("negative value in the vector: %f @ %d", v, i)
		}
	}
	for vi, j := range indices {
		if j >= wmh.dim {
			log.Panicf("index is out of range: %d @ %d", j, vi)
		}
	}
	hashvalues := make([]uint64, wmh.sampleSize)
	for s := 0; s < wmh.sampleSize; s++ {
		minLnA := float32(math.MaxFloat32)
		var k int
		var minT float32
		for vi, j := range indices {
			vlog := fastlog.Log(values[vi])
			beta := float32(wmh.betas[s][j]) / float32(maxUint16)
			// t = np.floor((vlog / self.rs[i]) + self.betas[i])
			t := float32(math.Floor(float64(vlog/wmh.rs[s][j] + beta)))
			// ln_y = (t - self.betas[i]) * self.rs[i]
			lnY := (t - beta) * wmh.rs[s][j]
			// ln_a = self.ln_cs[i] - ln_y - self.rs[i]
			lnA := wmh.lnCs[s][j] - lnY - wmh.rs[s][j]
			// k = np.nanargmin(ln_a)
			if lnA < minLnA {
				minLnA = lnA
				k = j
				minT = t
			}
		}
		// hashvalues[i][0], hashvalues[i][1] = k, int(t[k])
		switch wmh.Bitness {
		case 64:
			hashvalues[s] = uint64(uint64(k) | (uint64(minT) << 32))
		case 32:
			hashvalues[s] = uint64(uint32(k) | (uint32(minT) << 16))
		case 16:
			hashvalues[s] = uint64(uint16(k) | (uint16(minT) << 8))
		default:
			log.Fatalf("unsupported bitness value: %d", wmh.Bitness)
		}
	}
	return hashvalues
}
