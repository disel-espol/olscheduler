package balancer

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/disel-espol/olscheduler/balancer/worker"
	"github.com/disel-espol/olscheduler/httputil"
	"github.com/disel-espol/olscheduler/lambda"

	"github.com/lafikl/consistent"
)

func findWorkerNodeInSlice(workerNodeSlice []*worker.Node, target url.URL) int {
	totalItems := len(workerNodeSlice)
	for i := 0; i < totalItems; i++ {
		if workerNodeSlice[i].GetURL() == target {
			return i
		}
	}
	return -1
}

type PackageAware struct {
	hashRing      *consistent.Consistent
	loadThreshold uint
	workerNodes   []*worker.Node
	workerNodeMap map[string]*worker.Node
	mutex         *sync.Mutex
}

func getPotentialNode(largestPkg string, b *PackageAware) (*worker.Node, error) {

	salt := "salty"
	candidate1, err1 := b.hashRing.Get(largestPkg)
	candidate2, err2 := b.hashRing.Get(largestPkg + salt)
	if err1 != nil && err2 != nil {
		log.Fatal(err1, err2)
		return nil, errors.New("error in both candidates")
	} else if err1 != nil {
		return b.workerNodeMap[candidate2], nil
	} else if err2 != nil {
		return b.workerNodeMap[candidate1], nil
	}

	host1 := b.workerNodeMap[candidate1]
	host2 := b.workerNodeMap[candidate2]
	if host1.Load >= host2.Load {
		return host2, nil
	}

	return host1, nil

}

func (b *PackageAware) SelectWorker(r *http.Request, l *lambda.Lambda) (url.URL, *httputil.HttpError) {
	workerNodes := b.workerNodes
	if len(workerNodes) == 0 {
		return url.URL{}, httputil.New500Error("Can't select worker, Workers empty")
	}

	pkgs := l.Pkgs
	if len(pkgs) == 0 {
		return url.URL{}, httputil.New500Error("Can't select worker, No largest package, pkgs empty")
	}

	largestPkg := pkgs[0]

	selectedNode, err := getPotentialNode(largestPkg, b)
	if err != nil {
		log.Fatal(err)
		return url.URL{}, httputil.New500Error("Failed to select worker. URL not found")
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if selectedNode.Load >= b.loadThreshold { // Find least loaded
		selectedNode = b.selectLeastLoadedWorker()
	}

	selectedNode.Load++

	return selectedNode.GetURL(), nil

}

func (b *PackageAware) ReleaseWorker(workerUrl url.URL) {
	selectedNode := b.workerNodeMap[workerUrl.String()]
	if selectedNode != nil {
		b.mutex.Lock()
		defer b.mutex.Unlock()

		selectedNode.Load--
	}
}

func (b *PackageAware) AddWorker(workerUrl url.URL) {
	host := workerUrl.String()
	node := worker.NewNode(workerUrl)
	b.workerNodes = append(b.workerNodes, node)
	b.hashRing.Add(host)
	b.workerNodeMap[host] = node
}

func (b *PackageAware) RemoveWorker(workerUrl url.URL) {
	host := workerUrl.String()
	source := b.workerNodes
	targetIndex := findWorkerNodeInSlice(source, workerUrl)
	if targetIndex > -1 {
		b.workerNodes = append(source[:targetIndex], source[targetIndex+1:]...)
		b.hashRing.Remove(host)
		b.workerNodeMap[host] = nil
	}
}

func (b *PackageAware) selectLeastLoadedWorker() *worker.Node {
	targetIndex := 0
	workers := b.workerNodes
	for i := 1; i < len(workers); i++ {
		if workers[i].Load < workers[targetIndex].Load {
			targetIndex = i
		}
	}
	return workers[targetIndex]
}

func (b *PackageAware) GetAllWorkers() []url.URL {
	workerNodes := b.workerNodes
	totalWorkers := len(workerNodes)
	workerUrls := make([]url.URL, totalWorkers)

	for i, indexedNode := range workerNodes {
		workerUrls[i] = indexedNode.GetURL()
	}
	return workerUrls
}

func NewPackageAware(workerUrls []url.URL, loadThreshold uint) Balancer {
	totalUrls := len(workerUrls)

	workerNodes := make([]*worker.Node, totalUrls)
	workerNodeMap := make(map[string]*worker.Node)
	hashRing := consistent.New()

	for i, workerUrl := range workerUrls {
		urlString := workerUrl.String()
		workerNodes[i] = worker.NewNode(workerUrl)
		hashRing.Add(urlString)
		workerNodeMap[urlString] = workerNodes[i]
	}

	return &PackageAware{
		hashRing,
		loadThreshold,
		workerNodes,
		workerNodeMap,
		&sync.Mutex{}}
}

func NewPackageAwareFromJSONSlice(jsonSlice []string, loadThreshold uint) Balancer {
	return NewPackageAware(createWorkerURLSlice(jsonSlice), loadThreshold)
}
