package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func MakeRequest(url string, ch chan<- string) {
	start := time.Now()
	resp, _ := http.Get(url)

	secs := time.Since(start).Seconds()
	body, _ := ioutil.ReadAll(resp.Body)
	ch <- fmt.Sprintf("%.2f elapsed with response length: %d %s", secs, len(body), url)
}

func main() {
	start := time.Now()
	ch := make(chan string)
	port := 9090
	requests := []string{
		fmt.Sprintf("http://localhost:%d/status", port),
		fmt.Sprintf("http://localhost:%d/status", port),
		fmt.Sprintf("http://localhost:%d/status", port),
		fmt.Sprintf("http://localhost:%d/status", port),
		fmt.Sprintf("http://localhost:%d/status", port),
		fmt.Sprintf("http://localhost:%d/status", port),
		fmt.Sprintf("http://localhost:%d/status", port),
		fmt.Sprintf("http://localhost:%d/status", port),
	}
	for _, url := range requests {
		go MakeRequest(url, ch)
	}

	for range requests {
		fmt.Println(<-ch)
	}
	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}
