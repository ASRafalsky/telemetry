// Package compress contains wrappers for http.ResponseWriter and http.Request reader.
// It useful if you compress for some data on the flight in your middleware with need various type of compression.
// Enjoy!
package compress

import (
	"io"
	"net/http"

	"go.uber.org/multierr"
)

// CompressWriter describes CompressWriter.
type CompressWriter struct {
	w        http.ResponseWriter
	cw       writerCloser
	encoding string
}

type writerCloser interface {
	// Write writes data and set Content-Encoding to the header.
	Write(p []byte) (int, error)
	Close() error
}

// NewCompressWriter creates new CompressWriter instance.
func NewCompressWriter(w http.ResponseWriter, cw writerCloser, encoding string) *CompressWriter {
	return &CompressWriter{
		w:        w,
		cw:       cw,
		encoding: encoding,
	}
}

// Header returns http.Header.
func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

// Write writes data and sets Content-Encoding type to the header. It returns written bytes number and error if
// the Write failed.
func (c *CompressWriter) Write(p []byte) (int, error) {
	c.w.Header().Set("Content-Encoding", c.encoding)
	return c.cw.Write(p)
}

// WriteHeader sets status code to the Header.
func (c *CompressWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

// Close closes writer.
func (c *CompressWriter) Close() error {
	return c.cw.Close()
}

// CompressReader describes CompressReader.
type CompressReader struct {
	r  io.ReadCloser
	cr readerCloser
}

type readerCloser interface {
	Read(p []byte) (int, error)
	Close() error
}

// NewCompressReader creates new NewCompressReader instance.
func NewCompressReader(r io.ReadCloser, cr readerCloser) *CompressReader {
	return &CompressReader{
		r:  r,
		cr: cr,
	}
}

// Read reads data to the buffer. It returns read bytes number and error if
// the Read failed.
func (c *CompressReader) Read(p []byte) (int, error) {
	return c.cr.Read(p)
}

// Close closes everything inside the CompressReader and returns any errors if any occurred.
func (c *CompressReader) Close() error {
	err := c.r.Close()
	if errReader := c.cr.Close(); errReader != nil {
		err = multierr.Append(err, errReader)
	}
	return err
}
