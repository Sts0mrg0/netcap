/*
 * NETCAP - Traffic Analysis Framework
 * Copyright (c) 2017-2020 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

// Package io implements IO primitives
package io

import (
	"sync"

	"github.com/gogo/protobuf/proto"

	"github.com/dreadl0ck/netcap/delimited"
)

/*
 * Atomic Delimited Writer
 * A primitive for a concurrency safe writer for length delimited binary data
 */

// AtomicDelimitedWriter writes length delimited protobuf messages synchronized.
type AtomicDelimitedWriter struct {
	w delimited.Writer
	sync.Mutex
}

// PutProto writes a protocol buffer into the writer and returns an error.
func (a *AtomicDelimitedWriter) PutProto(pb proto.Message) error {
	a.Lock()
	err := a.w.PutProto(pb)
	a.Unlock()
	return err
}

// NewAtomicDelimitedWriter takes a delimited.WriterAtomic and returns an atomic version.
func NewAtomicDelimitedWriter(w *delimited.Writer) *AtomicDelimitedWriter {
	return &AtomicDelimitedWriter{
		w: *w,
	}
}
