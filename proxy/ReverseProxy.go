package proxy

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
	// worker node. It takes the same parameters as a HTTP server handler along
	// with the worker node as the first parameter. It is expected to send a
	// response to the client using the ResponseWriter object.
	ProxyRequest(workerURL url.URL, w http.ResponseWriter, r *http.Request)
}

type HTTPReverseProxy struct {
	proxyMap map[url.URL]*httputil.ReverseProxy
}

func (p *HTTPReverseProxy) getReverseProxyForWorker(workerURL url.URL) *httputil.ReverseProxy {
	proxy := p.proxyMap[workerURL]

	if proxy == nil {
		proxy = httputil.NewSingleHostReverseProxy(&workerURL)
		p.proxyMap[workerURL] = proxy
	}

	return proxy
}

func (p *HTTPReverseProxy) ProxyRequest(workerURL url.URL, w http.ResponseWriter, r *http.Request) {
	proxy := p.getReverseProxyForWorker(workerURL)
	proxy.ServeHTTP(w, r)
}

func NewHTTPReverseProxy() ReverseProxy {
	return &HTTPReverseProxy{make(map[url.URL]*httputil.ReverseProxy)}
}
