package service

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"hash"
	rand2 "math/rand"
)

type UUID [16]byte

const numBytes = "1234567890"

func getUUID() (uuidHex string) {
	uuid := newUUID()
	uuidHex = hex.EncodeToString(uuid[:])
	return
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = numBytes[rand2.Intn(len(numBytes))]
	}
	return string(b)
}

func newUUID() UUID {
	ns := UUID{}
	safeRandom(ns[:])
	u := newFromHash(md5.New(), ns, randStringBytes(16))
	u[6] = (u[6] & 0x0f) | (byte(2) << 4)
	u[8] = (u[8]&(0xff>>2) | (0x02 << 6))

	return u
}

func newFromHash(h hash.Hash, ns UUID, name string) UUID {
	u := UUID{}
	h.Write(ns[:])
	h.Write([]byte(name))
	copy(u[:], h.Sum(nil))

	return u
}

func safeRandom(dest []byte) {
	if _, err := rand.Read(dest); err != nil {
		panic(err)
	}
}
