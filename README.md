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

#### func  NewWithData

```go
func NewWithData(name string, data []byte) *File
```
NewWithData create a new File with the given name, and sets the initial
uncompressed data to that provided.

#### func (*File) Chtime

```go
func (f *File) Chtime(t time.Time)
```
Chtime sets the modtime to the given time.

#### func (*File) Create

```go
func (f *File) Create() *Writer
```
Create opens a Writer that can be used to write the data for the File. Close
must be called on the resulting Writer for the data to be accepted.

#### func (*File) Name

```go
func (f *File) Name() string
```
Name returns the name given during File creation.

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

#### func (*File) WriteTo

```go
func (f *File) WriteTo(w io.Writer) (int64, error)
```
WriteTo writes the uncompressed data to the given writer.

#### type Writer

```go
type Writer struct {
}
```

A Writer is bound to the File is was created from, buffering data that is
written to it.Writer. Upon Closing, that data will be compressed and both the
uncompressed and compressed data will be replaced on the File.

#### func (*Writer) Close

```go
func (f *Writer) Close() error
```
Close is an implementation of the io.Close interface.

This method must be called for the written data to be accepted on the File.

#### func (*Writer) Write

```go
func (f *Writer) Write(p []byte) (int, error)
```
Write is an implementation of the io.Writer interface.

#### func (*Writer) WriteString

```go
func (f *Writer) WriteString(str string) (int, error)
```
WriteString is an implementation of the io.StringWriter interface.
