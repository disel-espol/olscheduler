package httputil

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func ReplaceBodyWithString(r *http.Request, newStrBody string) {
	r.Body = ioutil.NopCloser(strings.NewReader(newStrBody))
	r.ContentLength = int64(len(newStrBody))
}

func GetBodyAsString(r *http.Request) string {
	body, _ := ioutil.ReadAll(r.Body)
	return string(body)
}

// Copied from OpenLamda src
// getUrlComponents parses request URL into its "/" delimated components
func GetUrlComponents(r *http.Request) []string {
	path := r.URL.Path
	// trim prefix
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	// trim trailing "/"
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	return strings.Split(path, "/")
}

func Get2ndPathSegment(r *http.Request, firstSegment string) string {
	components := GetUrlComponents(r)

	if len(components) != 2 {
		return ""
	}

	if components[0] != firstSegment {
		return ""
	}

	return components[1]
}
