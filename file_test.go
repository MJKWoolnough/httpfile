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

	testFile(t, server.URL, client, 1, http.MethodHead, "identity", "", earliest, latest, false)
	testFile(t, server.URL, client, 2, http.MethodGet, "identity", "", earliest, latest, false)

	earliest = time.Now().Add(-time.Second)
	f := file.Create()

	f.Write([]byte("some data"))
	io.WriteString(f, ", and some more data")
	f.Close()

	latest = time.Now().Add(time.Second)

	testFile(t, server.URL, client, 3, http.MethodGet, "identity", "some data, and some more data", earliest, latest, false)

	earliest = time.Now().Add(-time.Second)
	f = file.Create()

	f.WriteString("checking compressed data")
	f.Close()

	latest = time.Now().Add(time.Second)

	testFile(t, server.URL, client, 4, http.MethodGet, "", "checking compressed data", earliest, latest, true)
}

func testFile(t *testing.T, url string, client *http.Client, test int, method, encoding, data string, earliest, latest time.Time, compressed bool) {
	var (
		r   *http.Response
		err error
	)

	if encoding != "" {
		req, _ := http.NewRequest(method, url, nil)

		req.Header.Set("Accept-Encoding", encoding)

		r, err = client.Do(req)
	} else {
		r, err = client.Get(url)
	}

	if err != nil {
		t.Fatalf("test %d: unexpected error: %s", test, err)
	}

	if encoding != "" {
		if r.ContentLength != int64(len(data)) {
			t.Errorf("test %d: expecting to have size %d, got %d", test, len(data), r.ContentLength)
		}
	}

	if mime := r.Header.Get("Content-Type"); mime != "application/json" {
		t.Errorf("test %d: expecting to have json mime type, got %q", test, mime)
	}

	if r.Uncompressed != compressed {
		t.Errorf("test %d: expecting Uncompressed to be %v, got %v", test, r.Uncompressed, compressed)
	}

	if modtime, err := time.Parse(time.RFC1123, r.Header.Get("Last-Modified")); err != nil {
		t.Errorf("test %d: error parsing modtime: %s", test, err)
	} else if earliest.After(modtime) || modtime.After(latest) {
		t.Errorf("test %d: expecting modtime to be between %s and %s, got %s", test, earliest, latest, modtime)
	}

	var sb strings.Builder

	io.Copy(&sb, r.Body)

	if sb.String() != data {
		t.Errorf("test %d: expecting to read content %q, got %q", test, data, sb.String())
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
