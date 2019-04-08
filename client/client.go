package client

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type httpReq struct {
	path  string
	query string
	port  int
}

func sendRequest(httpReq httpReq) (int, error) {
	httpUrl := fmt.Sprintf("http://localhost:%d%s?%s", httpReq.port,
		httpReq.path, httpReq.query)
	resp, err := http.Post(httpUrl, "", nil)

	if err != nil {
		return 0, nil
	}

	return resp.StatusCode, nil
}

func encodeWorkerUrls(urls []string) string {
	v := url.Values{}
	for _, workerUrl := range urls {
		v.Add("workers", workerUrl)
	}
	return v.Encode()
}

func AddWorkers(port int, workerUrls []string) error {
	req := httpReq{
		query: encodeWorkerUrls(workerUrls),
		port:  port,
		path:  "/admin/workers/add",
	}

	status, err := sendRequest(req)
	if err != nil {
		return err
	}

	if status != 200 {
		msg := fmt.Sprintf("Unexpected error code: %d", status)
		return errors.New(msg)
	}

	return nil
}

func RemoveWorkers(port int, workerUrls []string) error {
	req := httpReq{
		query: encodeWorkerUrls(workerUrls),
		port:  port,
		path:  "/admin/workers/remove",
	}

	status, err := sendRequest(req)
	if err != nil {
		return err
	}

	if status != 200 {
		msg := fmt.Sprintf("Unexpected error code: %d", status)
		return errors.New(msg)
	}

	return nil
}
