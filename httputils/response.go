package httputils

import (
	"net/http"
	"strings"
)

type Response struct {
	RawResponse *http.Response
	Header      http.Header
	Body        []byte
	StatusCode  int
}

func (res *Response) String() string {
	return strings.TrimSpace(string(res.Body))
}
