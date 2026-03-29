package worker

import (
	"io"
	"mime/multipart"
	"net/http"
)

// multipartWriter wraps multipart.Writer for testability.
type multipartWriter struct {
	*multipart.Writer
}

func newMultipartWriter(w io.Writer) *multipartWriter {
	return &multipartWriter{multipart.NewWriter(w)}
}

// httpPost is a simple wrapper around http.Post.
func httpPost(url, contentType string, body io.Reader) (*http.Response, error) {
	return http.Post(url, contentType, body)
}
