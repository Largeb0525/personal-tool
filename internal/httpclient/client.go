package httpclient

import (
	"net/http"
	"time"
)

var (
	// DefaultClient is the shared HTTP client for the entire application.
	DefaultClient *http.Client
)

func init() {
	DefaultClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}
