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

var empty = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\x03\x03\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")

// Type file represents an http.Handler upon which you can set static data.
type File struct {
	name string

	mu               sync.RWMutex
	modtime          time.Time
	data, compressed []byte
}

// New creates a new File with the given name, which is used to apply Content-Type headers.
func New(name string) *File {
	return &File{name: name, modtime: time.Now(), compressed: empty}
}

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

func (f *File) WriteTo(w io.Writer) (int64, error) {
	f.mu.RLock()
	data := f.data
	f.mu.RUnlock()

	n, err := w.Write(data)

	return int64(n), err
}

// Chtime sets the motime to the given time.
func (f *File) Chtime(t time.Time) {
	f.mu.Lock()
	f.modtime = t
	f.mu.Unlock()
}

func (f *File) Name() string {
	return f.name
}

type file struct {
	file *File
	data []byte
}

type Writer interface {
	io.WriteCloser
	io.StringWriter
}

func (f *File) Create() Writer {
	return &file{
		file: f,
	}
}

func (f *file) Write(p []byte) (int, error) {
	if f.file == nil {
		return 0, fs.ErrClosed
	}

	f.data = append(f.data, p...)

	return len(p), nil
}

func (f *file) WriteString(str string) (int, error) {
	if f.file == nil {
		return 0, fs.ErrClosed
	}

	f.data = append(f.data, str...)

	return len(str), nil
}

func (f *file) Close() error {
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

	*f = file{}

	return nil
}
