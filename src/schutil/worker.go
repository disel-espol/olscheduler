package schutil

import (
	"net/url"
	"net/http/httputil"
)

type Worker struct {
	URL *url.URL
	ReverseProxy *httputil.ReverseProxy
	Load int
	Weight int
}