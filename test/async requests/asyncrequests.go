package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func MakeRequest(url string, body []byte, ch chan<- string) {
	start := time.Now()
	resp, _ := http.Post(url, "application/json", bytes.NewBuffer(body))

	secs := time.Since(start).Seconds()
	b, _ := ioutil.ReadAll(resp.Body)
	ch <- fmt.Sprintf("%.2f elapsed with response length: %d %s", secs, len(b), url)
}

func main() {
	start := time.Now()
	ch := make(chan string)
	port := 9090
	lambda := "hello"
	body := []byte(`{"pkgs": ["fmt", "rand"], "name": "Moon"}`)
	requests := []string{
		fmt.Sprintf("http://localhost:%d/runLambda/%s", port, lambda),
		fmt.Sprintf("http://localhost:%d/runLambda/%s", port, lambda),
		fmt.Sprintf("http://localhost:%d/runLambda/%s", port, lambda),
		fmt.Sprintf("http://localhost:%d/runLambda/%s", port, lambda),
		fmt.Sprintf("http://localhost:%d/runLambda/%s", port, lambda),
		fmt.Sprintf("http://localhost:%d/runLambda/%s", port, lambda),
		fmt.Sprintf("http://localhost:%d/runLambda/%s", port, lambda),
		fmt.Sprintf("http://localhost:%d/runLambda/%s", port, lambda),
	}
	for range [2]struct{}{} {
		requests = append(requests, requests...)
	}

	for _, url := range requests {
		go MakeRequest(url, body, ch)
		time.Sleep(500 * time.Microsecond)
	}

	for range requests { // Awaiting all responses
		fmt.Println(<-ch)
	}
	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}
