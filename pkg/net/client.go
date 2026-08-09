package net

import (
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	defaultClient     *http.Client
	defaultClientOnce sync.Once
)

// DefaultHTTPClient returns a thread-safe, connection-pooled HTTP client.
func DefaultHTTPClient() *http.Client {
	defaultClientOnce.Do(func() {
		defaultClient = &http.Client{
			// Timeout intentionally left at zero (unset).
			// Context-based cancellation is the only timeout mechanism.
			// A hardcoded HTTP-level deadline would prematurely kill long-running
			// SSE model-streaming calls before the request has a chance to complete.
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   10,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
	})
	return defaultClient
}
