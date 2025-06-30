package api

import (
	"log/slog" // Импортируем slog
	"net/http"
	"net/http/httputil"
)

type loggingTransport struct {
	Logger    *slog.Logger
	Transport http.RoundTripper
}

func (lt *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	_, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		lt.Logger.Error("Error dumping request", "error", err)
		return nil, err
	}

	resp, err := lt.Transport.RoundTrip(req)
	if err != nil {
		lt.Logger.Error("Error performing request", "error", err)
		return nil, err
	}

	_, err = httputil.DumpResponse(resp, true)
	if err != nil {
		lt.Logger.Error("Error dumping response", "error", err)
		return nil, err
	}

	lt.Logger.Info("HTTP Request/Response",
		slog.String("url", req.URL.String()),
		//slog.String("request", string(reqDump)),
		//slog.String("response", string(respDump)),
	)

	return resp, nil
}
