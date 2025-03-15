# httpfile
--
    import "vimagination.zapto.org/httpfile"

Package httpfile provides an easy way to create HTTP handlers that respond with
static data, possibly gzip compressed if requested by the client.

## Usage

#### type File

```go
type File struct {
}
```

Type file represents an http.Handler upon which you can set static data.

#### func  New

```go
func New(name string) *File
```
New creates a new File with the given name, which is used to apply Content-Type
headers.

#### func (*File) Chtime

```go
func (f *File) Chtime(t time.Time)
```
Chtime sets the motime to the given time.

#### func (*File) Create

```go
func (f *File) Create() Writer
```

#### func (*File) ReadFrom

```go
func (f *File) ReadFrom(r io.Reader) (int64, error)
```
ReadFrom reads all of the data from io.Reader and applies it to the file,
overwriting any existing data and setting the modtime to Now.

#### func (*File) ServeHTTP

```go
func (f *File) ServeHTTP(w http.ResponseWriter, r *http.Request)
```
ServeHTTP implements the http.Handler interface.

#### type Writer

```go
type Writer interface {
	io.WriteCloser
	io.StringWriter
}
```
