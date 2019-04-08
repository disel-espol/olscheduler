package balancer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"../httputil"
	"../lambda"
	"../worker"

	"github.com/lafikl/consistent"
)

type PkgAwareBalancer struct {
	loadThreshold int
	c             *consistent.Consistent
	m             map[string]int
}

func (b *PkgAwareBalancer) Init(workers []string, threshold int) {
	b.c = consistent.New()
	b.loadThreshold = threshold
	b.m = make(map[string]int)
	for i := 0; i < len(workers); i = i + 2 {
		host := "http://" + workers[i]
		b.m[host] = i / 2
		b.c.Add(host)
		fmt.Println(host)
	}

}

func stripPkgsArrayFromBody(strBody string) ([]string, string, *httputil.HttpError) {
	pkgsRegExp := regexp.MustCompile(`"pkgs"\s*:\s*\[.*\],*\s*`)
	matches := pkgsRegExp.FindStringSubmatch(strBody)
	if len(matches) < 1 {
		return nil, "", httputil.New400Error("Pkgs array required")
	}

	strPkgsJson := matches[0]
	pkgsArrayRegExp := regexp.MustCompile(`\[.*\]`)
	srtPkgsMatches := pkgsArrayRegExp.FindStringSubmatch(strPkgsJson)
	if len(srtPkgsMatches) < 1 {
		return nil, "", httputil.New400Error("Pkgs array ill-formed")
	}

	decoder := json.NewDecoder(strings.NewReader(srtPkgsMatches[0]))
	var pkgs []string // pkgs ordered from larger to smaller
	err := decoder.Decode(&pkgs)
	if err != nil {
		return nil, "", httputil.New400Error("Malformed JSON string")
	}

	newStrBody := strings.Replace(strBody, strPkgsJson, "", -1)

	return pkgs, newStrBody, nil
}

func (b *PkgAwareBalancer) selectWorkerPkgAware(
	workers []*worker.Worker,
	pkgs []string,
	threshold int) (*worker.Worker, *httputil.HttpError) {
	if len(workers) == 0 {
		return nil, httputil.New500Error("Can't select worker, Workers empty")
	}

	if len(pkgs) == 0 {
		return nil, httputil.New500Error("Can't select worker, No largest package, pkgs empty")
	}

	largestPkg := pkgs[0]
	host, err := b.c.Get(largestPkg)
	if err != nil {
		log.Fatal(err)
	}
	targetIndex := b.m[host]

	if workers[targetIndex].GetLoad() >= threshold { // Find least loaded
		targetIndex = 0
		for i := 1; i < len(workers); i++ {
			if workers[i].GetLoad() < workers[targetIndex].GetLoad() {
				targetIndex = i
			}
		}
	}

	return workers[targetIndex], nil
}

func (b *PkgAwareBalancer) SelectWorker(
	workers []*worker.Worker,
	r *http.Request,
	l *lambda.Lambda) (*worker.Worker, *httputil.HttpError) {
	return b.selectWorkerPkgAware(workers, l.Pkgs, b.loadThreshold)
}
