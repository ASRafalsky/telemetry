package compress

import (
	"io"
	"net/http"

	"go.uber.org/multierr"
)

type CompressWriter struct {
	w        http.ResponseWriter
	cw       writerCloser
	encoding string
}

type writerCloser interface {
	Write(p []byte) (int, error)
	Close() error
}

func NewCompressWriter(w http.ResponseWriter, cw writerCloser, encoding string) *CompressWriter {
	return &CompressWriter{
		w:        w,
		cw:       cw,
		encoding: encoding,
	}
}

func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	c.w.Header().Set("Content-Encoding", c.encoding)
	return c.cw.Write(p)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	return c.cw.Close()
}

type CompressReader struct {
	r  io.ReadCloser
	cr readerCloser
}

type readerCloser interface {
	Read(p []byte) (int, error)
	Close() error
}

func NewCompressReader(r io.ReadCloser, cr readerCloser) *CompressReader {
	return &CompressReader{
		r:  r,
		cr: cr,
	}
}

func (c *CompressReader) Read(p []byte) (int, error) {
	return c.cr.Read(p)
}

func (c *CompressReader) Close() error {
	err := c.r.Close()
	if errReader := c.cr.Close(); errReader != nil {
		err = multierr.Append(err, errReader)
	}
	return err
}
