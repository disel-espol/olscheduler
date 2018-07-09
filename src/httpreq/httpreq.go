package httpreq

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
