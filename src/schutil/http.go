package schutil

import (
	"net/http"
)

type HttpError struct {
	Msg  string
	Code int
}

type AppendResponseWriter struct {
	headers   http.Header
	Body      []byte
	Status    int
	separator []byte
}

func NewAppendResponseWriter() *AppendResponseWriter {
	return &AppendResponseWriter{headers: make(http.Header), separator: []byte("\n")}
}

func (this *AppendResponseWriter) Header() http.Header {
	return this.headers
}

func (this *AppendResponseWriter) Write(body []byte) (int, error) {
	if len(this.Body) > 0 {
		this.Body = append(this.Body, this.separator...)
	}
	this.Body = append(this.Body, body...)
	return len(this.Body), nil
}

func (this *AppendResponseWriter) WriteHeader(status int) {
	this.Status = status
}

type ObserverResponseWriter struct {
	Body   []byte
	Status int

	rw http.ResponseWriter
}

func NewObserverResponseWriter(rw http.ResponseWriter) *ObserverResponseWriter {
	return &ObserverResponseWriter{rw: rw}
}

func (this *ObserverResponseWriter) Header() http.Header {
	return this.rw.Header()
}

func (this *ObserverResponseWriter) Write(body []byte) (int, error) {
	this.Body = append(this.Body, body...)
	return this.rw.Write(body)
}

func (this *ObserverResponseWriter) WriteHeader(status int) {
	this.Status = status
	this.rw.WriteHeader(status)
}
