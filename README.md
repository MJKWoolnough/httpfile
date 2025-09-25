# httpfile

[![CI](https://github.com/MJKWoolnough/httpfile/actions/workflows/go-checks.yml/badge.svg)](https://github.com/MJKWoolnough/httpfile/actions)
[![Go Reference](https://pkg.go.dev/badge/vimagination.zapto.org/httpfile.svg)](https://pkg.go.dev/vimagination.zapto.org/httpfile)
[![Go Report Card](https://goreportcard.com/badge/vimagination.zapto.org/httpfile)](https://goreportcard.com/report/vimagination.zapto.org/httpfile)

--
    import "vimagination.zapto.org/httpfile"

Package httpfile provides an easy way to create HTTP handlers that respond with static data, possibly gzip compressed if requested by the client.

## Highlights

 - Create an updatable `http.Handler` from static data.
 - Sends either gzip-compressed or uncompressed data based on `Accept-Encoding` header.

## Usage

```go
package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"

	"vimagination.zapto.org/httpfile"
)

func main() {
	handler := httpfile.NewWithData("file.json", []byte(`{"hello", "world!"}`))

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-encoding", "identity")

	handler.ServeHTTP(w, r)

	fmt.Println(w.Body)

	handler.ReadFrom(bytes.NewBuffer([]byte(`{"foo": "bar"}`)))

	r.Header.Set("Accept-encoding", "identity")

	w = httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	fmt.Println(w.Body)

	// Output:
	// {"hello", "world!"}
	// {"foo": "bar"}
}
}
```

## Documentation

Full API docs can be found at:

https://pkg.go.dev/vimagination.zapto.org/httpfile
