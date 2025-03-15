package httpfile

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFile(t *testing.T) {
	earliest := time.Now().Add(-time.Second)
	file := New("data.json")
	latest := time.Now().Add(time.Second)

	server := httptest.NewServer(file)
	defer server.Close()

	client := server.Client()

	r, err := client.Head(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if r.ContentLength != 0 {
		t.Errorf("expecting to have no size, got %d", r.ContentLength)
	}

	if mime := r.Header.Get("Content-Type"); mime != "application/json" {
		t.Errorf("expecting to have json mime type, got %q", mime)
	}

	if modtime, err := time.Parse(time.RFC1123, r.Header.Get("Last-Modified")); err != nil {
		t.Errorf("error parsing modtime: %s", err)
	} else if earliest.After(modtime) || modtime.After(latest) {
		t.Errorf("expecting modtime to be between %s and %s, got %s", earliest, latest, modtime)
	}

	earliest = time.Now().Add(-time.Second)
	f := file.Create()

	f.Write([]byte("some data"))
	io.WriteString(f, ", and some more data")
	f.Close()

	latest = time.Now().Add(time.Second)

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	req.Header.Set("Accept-Encoding", "")

	r, err = client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if r.Uncompressed != false {
		t.Errorf("expecting to receive uncompressed data, didn't!")
	}

	if r.ContentLength != 29 {
		t.Errorf("expecting to have size of 29 bytes, got %d", r.ContentLength)
	}

	var sb strings.Builder

	io.Copy(&sb, r.Body)

	if sb.String() != "some data, and some more data" {
		t.Errorf("expecting to read content %q, got %q", "some data, and some more data", sb.String())
	}

	if modtime, err := time.Parse(time.RFC1123, r.Header.Get("Last-Modified")); err != nil {
		t.Errorf("error parsing modtime: %s", err)
	} else if earliest.After(modtime) || modtime.After(latest) {
		t.Errorf("expecting modtime to be between %s and %s, got %s", earliest, latest, modtime)
	}

	earliest = time.Now().Add(-time.Second)
	f = file.Create()

	f.WriteString("checking compressed data")
	f.Close()

	latest = time.Now().Add(time.Second)

	r, err = client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if r.Uncompressed != true {
		t.Errorf("expecting to receive compressed data, didn't!")
	}

	if modtime, err := time.Parse(time.RFC1123, r.Header.Get("Last-Modified")); err != nil {
		t.Errorf("error parsing modtime: %s", err)
	} else if earliest.After(modtime) || modtime.After(latest) {
		t.Errorf("expecting modtime to be between %s and %s, got %s", earliest, latest, modtime)
	}

	sb.Reset()

	io.Copy(&sb, r.Body)

	if sb.String() != "checking compressed data" {
		t.Errorf("expecting to read content %q, got %q", "checking compressed data", sb.String())
	}
}

func TestReadFrom(t *testing.T) {
	const data = "<html><head><body><h1>Hello, World</h1></body></html>"

	file := New("data.html")

	server := httptest.NewServer(file)
	defer server.Close()

	n, _ := file.ReadFrom(strings.NewReader(data))
	if n != 53 {
		t.Errorf("expecting to write 53 bytes, wrote %d", n)
	}

	r, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	var sb strings.Builder

	io.Copy(&sb, r.Body)

	if sb.String() != data {
		t.Errorf("expecting to read content %q, got %q", data, sb.String())
	}
}

func TestChtime(t *testing.T) {
	expected := time.Unix(12345, 0)
	file := New("data.txt")

	file.Chtime(expected)

	server := httptest.NewServer(file)
	defer server.Close()

	r, err := server.Client().Head(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if modtime, err := time.Parse(time.RFC1123, r.Header.Get("Last-Modified")); err != nil {
		t.Errorf("error parsing modtime: %s", err)
	} else if !modtime.Equal(expected) {
		t.Errorf("expecting modtime to be %s, got %s", expected, modtime)
	}
}
