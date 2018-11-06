package worker

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ReverseProxy is an interface for an object that proxies the client HTTP
// request to a worker node. The implementation is free to choose any
// protocol or network stack
type ReverseProxy interface {
	// ProxyRequest is a function that proxies the client HTTP request to a
	// worker node. It takes the same parameters as a HTTP server handler.
	// It is expected to send a response to the client using the
	// ResponseWriter object.
	ProxyRequest(w http.ResponseWriter, r *http.Request)
}

type HTTPReverseProxy struct {
	handler *httputil.ReverseProxy
}

func NewHTTPReverseProxy(u *url.URL) *HTTPReverseProxy {
	return &HTTPReverseProxy{handler: httputil.NewSingleHostReverseProxy(u)}
}

func (p *HTTPReverseProxy) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	p.handler.ServeHTTP(w, r)
}
