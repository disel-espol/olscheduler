package schutil

import (
	"net/http/httputil"
	"net/url"
)

type Worker struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
	Load         int
	Weight       int
}
