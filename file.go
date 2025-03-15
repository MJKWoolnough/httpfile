package httpfile

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

type File struct {
	name string

	mu               sync.RWMutex
	modtime          time.Time
	data, compressed []byte
}

func New(name string) *File {
	return &File{name: name, modtime: time.Now()}
}

var isGzip = httpencoding.HandlerFunc(func(enc httpencoding.Encoding) bool { return enc == "gzip" })

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

func (f *File) ReadFrom(r io.Reader) (int64, error) {
	file := f.Create()
	defer file.Close()

	return io.Copy(file, r)
}

func (f *File) Chtime(t time.Time) {
	f.mu.Lock()
	f.modtime = t
	f.mu.Unlock()
}

type file struct {
	file *File
	data []byte
}

func (f *File) Create() io.WriteCloser {
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
