package balancer

import (
	"encoding/json"
	"hash/fnv"
	"net/http"
	"regexp"
	"strings"

	"../httputil"
	"../lambda"
	"../worker"
)

type PkgAwareBalancer struct {
	loadThreshold int
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

func h1(s string) uint32 {
	hf := fnv.New32()
	hf.Write([]byte(s))
	return hf.Sum32()
}

func h2(s string) uint32 {
	hf := fnv.New32a()
	hf.Write([]byte(s))
	return hf.Sum32()
}

func selectWorkerPkgAware(workers []*worker.Worker,
	pkgs []string,
	threshold int) (*worker.Worker, *httputil.HttpError) {
	if len(workers) == 0 {
		return nil, httputil.New500Error("Can't select worker, Workers empty")
	}

	if len(pkgs) == 0 {
		return nil, httputil.New500Error("Can't select worker, No largest package, pkgs empty")
	}

	largestPkg := pkgs[0]
	targetIndex1 := h1(largestPkg) % uint32(len(workers))
	targetIndex2 := h2(largestPkg) % uint32(len(workers))

	targetIndex := targetIndex2
	if workers[targetIndex1].GetLoad() < workers[targetIndex2].GetLoad() {
		targetIndex = targetIndex1
	}

	if workers[targetIndex].GetLoad() >= threshold { // Find least loaded
		targetIndex = 0
		for i := 1; i < len(workers); i++ {
			if workers[i].GetLoad() < workers[targetIndex].GetLoad() {
				targetIndex = uint32(i)
			}
		}
	}

	return workers[targetIndex], nil
}

func (b *PkgAwareBalancer) SelectWorker(workers []*worker.Worker, r *http.Request, l *lambda.Lambda) (*worker.Worker, *httputil.HttpError) {
	return selectWorkerPkgAware(workers, l.Pkgs, b.loadThreshold)
}
