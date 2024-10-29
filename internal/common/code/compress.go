package code

import (
	"compress/gzip"
	"github.com/andybalholm/brotli"
	"io"
)

type writeCloserWrapper struct {
	io.Writer
}

func (w *writeCloserWrapper) Close() error {
	// No-op close method
	return nil
}

func WarpReader(reader io.ReadCloser, compressType string) (r io.ReadCloser, err error) {
	switch compressType {
	case "gzip":
		r, err = gzip.NewReader(reader)
	case "br":
		r = io.NopCloser(brotli.NewReader(reader))
	default:
		r = reader
	}
	return r, err
}

func WarpWriter(writer io.Writer, compressType string) (w io.WriteCloser, err error) {
	switch compressType {
	case "gzip":
		return gzip.NewWriter(writer), nil
	case "br":
		return brotli.NewWriter(writer), nil
	default:
		return &writeCloserWrapper{Writer: writer}, nil
	}
}
