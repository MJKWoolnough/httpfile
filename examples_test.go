package httpfile_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"

	"vimagination.zapto.org/httpfile"
)

func Example() {
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
