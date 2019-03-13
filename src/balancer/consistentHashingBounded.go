package balancer

import (
	"net/http"
	"fmt"
	"log"
	"../httputil"
	"../lambda"
	"../worker"
	"github.com/lafikl/consistent"
)

type ConsistentHashingBounded struct {
	c *consistent.Consistent
	m map[string]int
}

func (b *ConsistentHashingBounded) init(workers []string){
	b.c=consistent.New()
	b.m=make(map[string]int)
	for i:=0; i < len(workers); i = i+ 2 {
		host:= "http://" + workers[i]
		b.m[host] = i/2
		b.c.Add(host)
		fmt.Println(host)
	}
}


func (b *ConsistentHashingBounded) SelectWorker(workers []*worker.Worker, r *http.Request, l *lambda.Lambda) (*worker.Worker, *httputil.HttpError) {
	if len(workers) == 0 {
		return nil, httputil.New500Error("Can't select worker, Workers empty")
	}

	host, err := b.c.GetLeast(r.URL.String())
	if err != nil {
		log.Fatal(err)
	}
	b.c.Inc(host)
	return workers[b.m[host]], nil
}

func (b *ConsistentHashingBounded) ReleaseWorker(host string){
	b.c.Done(host)
}

