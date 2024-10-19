package api

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

type loggingTransport struct{}

func (s *loggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	b, _ := httputil.DumpRequestOut(r, true)

	resp, err := http.DefaultTransport.RoundTrip(r)
	// err is returned after dumping the response

	respBytes, _ := httputil.DumpResponse(resp, true)
	b = append(b, respBytes...)

	fmt.Printf("%s\n", b)

	return resp, err
}
