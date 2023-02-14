// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package pargzip contains a parallel gzip writer implementation.  By
// compressing each chunk of data in parallel, all the CPUs on the
// machine can be used, at a slight loss of compression efficiency.
package pargzip

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"runtime"
	"strings"
	"sync"
)

// A Writer is an io.WriteCloser.
// Writes to a Writer are compressed and written to w.
//
// Any exported fields may only be mutated before the first call to
// Write.
type Writer struct {
	// ChunkSize is the number of bytes to gzip at once.
	// The default from NewWriter is 1MB.
	ChunkSize int

	// Parallel is the number of chunks to compress in parallel.
	// The default from NewWriter is runtime.NumCPU().
	Parallel int

	w  io.Writer
	bw *bufio.Writer

	allWritten  chan struct{} // when writing goroutine ends
	wasWriteErr chan struct{} // closed after 'err' set

	sem    chan bool        // semaphore bounding compressions in flight
	chunkc chan *writeChunk // closed on Close

	mu     sync.Mutex // guards following
	closed bool
	err    error // sticky write error
}

type writeChunk struct {
	zw *Writer
	p  string // uncompressed

	donec chan struct{} // closed on completion

	// one of following is set:
	z   []byte // compressed
	err error  // exec error
}

// compress runs the gzip child process.
// It runs in its own goroutine.
func (c *writeChunk) compress() (err error) {
	defer func() {
		if err != nil {
			c.err = err
		}
		close(c.donec)
		<-c.zw.sem
	}()
	var zbuf bytes.Buffer
	zw := gzip.NewWriter(&zbuf)
	if _, err := io.Copy(zw, strings.NewReader(c.p)); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	c.z = zbuf.Bytes()
	return nil
}

// NewWriter returns a new Writer.
// Writes to the returned writer are compressed and written to w.
//
// It is the caller's responsibility to call Close on the WriteCloser
// when done. Writes may be buffered and not flushed until Close.
//
// Any fields on Writer may only be modified before the first call to
// Write.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:           w,
		allWritten:  make(chan struct{}),
		wasWriteErr: make(chan struct{}),

		ChunkSize: 1 << 20,
		Parallel:  runtime.NumCPU(),
	}
}

func (w *Writer) didInit() bool { return w.bw != nil }

func (w *Writer) init() {
	w.bw = bufio.NewWriterSize(newChunkWriter{w}, w.ChunkSize)
	w.chunkc = make(chan *writeChunk, w.Parallel+1)
	w.sem = make(chan bool, w.Parallel)
	go func() {
		defer close(w.allWritten)
		for c := range w.chunkc {
			if err := w.writeCompressedChunk(c); err != nil {
				close(w.wasWriteErr)
				return
			}
		}
	}()
}

func (w *Writer) startChunk(p []byte) {
	w.sem <- true // block until we can begin
	c := &writeChunk{
		zw:    w,
		p:     string(p), // string, since the bufio.Writer owns the slice
		donec: make(chan struct{}),
	}
	go c.compress() // receives from w.sem
	select {
	case w.chunkc <- c:
	case <-w.wasWriteErr:
		// Discard chunks that come after any chunk that failed
		// to write.
	}
}

func (w *Writer) writeCompressedChunk(c *writeChunk) (err error) {
	defer func() {
		if err != nil {
			w.mu.Lock()
			defer w.mu.Unlock()
			if w.err == nil {
				w.err = err
			}
		}
	}()
	<-c.donec
	if c.err != nil {
		return c.err
	}
	_, err = w.w.Write(c.z)
	return
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if !w.didInit() {
		w.init()
	}
	return w.bw.Write(p)
}

func (w *Writer) Close() error {
	w.mu.Lock()
	err, wasClosed := w.err, w.closed
	w.closed = true
	w.mu.Unlock()
	if wasClosed {
		return nil
	}
	if !w.didInit() {
		return nil
	}
	if err != nil {
		return err
	}

	w.bw.Flush()
	close(w.chunkc)
	<-w.allWritten // wait for writing goroutine to end

	w.mu.Lock()
	err = w.err
	w.mu.Unlock()
	return err
}

// newChunkWriter gets large chunks to compress and write to zw.
type newChunkWriter struct {
	zw *Writer
}

func (cw newChunkWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	max := cw.zw.ChunkSize
	for len(p) > 0 {
		chunk := p
		if len(chunk) > max {
			chunk = chunk[:max]
		}
		p = p[len(chunk):]
		cw.zw.startChunk(chunk)
	}
	return
}
