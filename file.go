// Package httpfile provides an easy way to create HTTP handlers that respond with static data, possibly gzip compressed if requested by the client.
package httpfile // import "vimagination.zapto.org/httpfile"

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"vimagination.zapto.org/httpencoding"
)

var empty = [20]byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x03, 0x03}

// Type file represents an http.Handler upon which you can set static data.
type File struct {
	name string

	mu               sync.RWMutex
	modtime          time.Time
	data, compressed []byte
}

// New creates a new File with the given name, which is used to apply
// Content-Type headers.
func New(name string) *File {
	return &File{name: name, modtime: time.Now(), compressed: empty[:]}
}

// NewWithData create a new File with the given name, and sets the initial
// uncompressed data to that provided.
func NewWithData(name string, data []byte) *File {
	var buf bytes.Buffer

	f := New(name)
	f.data = data
	g := gzip.NewWriter(&buf)

	g.Write(data)
	g.Close()

	f.compressed = buf.Bytes()

	return f
}

var isGzip = httpencoding.HandlerFunc(func(enc httpencoding.Encoding) bool { return enc == "gzip" })

// ServeHTTP implements the http.Handler interface.
func (f *File) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Reader

	wantsGzip := httpencoding.HandleEncoding(r, isGzip)

	f.mu.RLock()

	modtime := f.modtime

	if wantsGzip {
		buf.Reset(f.compressed)
		w.Header().Add("Content-Encoding", "gzip")
	} else {
		buf.Reset(f.data)
	}

	f.mu.RUnlock()

	http.ServeContent(w, r, f.name, modtime, &buf)
}

// ReadFrom reads all of the data from io.Reader and applies it to the file,
// overwriting any existing data and setting the modtime to Now.
func (f *File) ReadFrom(r io.Reader) (int64, error) {
	file := f.Create()
	defer file.Close()

	return io.Copy(file, r)
}

// WriteTo writes the uncompressed data to the given writer.
func (f *File) WriteTo(w io.Writer) (int64, error) {
	f.mu.RLock()
	data := f.data
	f.mu.RUnlock()

	n, err := w.Write(data)

	return int64(n), err
}

// Chtime sets the modtime to the given time.
func (f *File) Chtime(t time.Time) {
	f.mu.Lock()
	f.modtime = t
	f.mu.Unlock()
}

// Name returns the name given during File creation.
func (f *File) Name() string {
	return f.name
}

// A Writer is bound to the File is was created from, buffering data that is
// written to it. Upon Closing, that data will be compressed and both the
// uncompressed and compressed data will be replaced on the File.
type Writer struct {
	file *File
	data []byte
}

// Create opens a Writer that can be used to write the data for the File. Close
// must be called on the resulting Writer for the data to be accepted.
func (f *File) Create() *Writer {
	return &Writer{
		file: f,
	}
}

// Write is an implementation of the io.Writer interface.
func (f *Writer) Write(p []byte) (int, error) {
	if f.file == nil {
		return 0, fs.ErrClosed
	}

	f.data = append(f.data, p...)

	return len(p), nil
}

// WriteString is an implementation of the io.StringWriter interface.
func (f *Writer) WriteString(str string) (int, error) {
	if f.file == nil {
		return 0, fs.ErrClosed
	}

	f.data = append(f.data, str...)

	return len(str), nil
}

// Close is an implementation of the io.Close interface.
//
// This method must be called for the written data to be accepted on the File.
func (f *Writer) Close() error {
	if f.file == nil {
		return fs.ErrClosed
	}

	var compressed bytes.Buffer

	g := gzip.NewWriter(&compressed)

	g.Write(f.data)
	g.Close()

	f.file.mu.Lock()

	f.file.data = f.data
	f.file.compressed = compressed.Bytes()
	f.file.modtime = time.Now()

	f.file.mu.Unlock()

	*f = Writer{}

	return nil
}
