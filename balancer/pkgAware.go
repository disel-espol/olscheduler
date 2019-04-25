package balancer

import (
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
	host, err := b.hashRing.Get(largestPkg)
	if err != nil {
		log.Fatal(err)
	}
	selectedNode := b.workerNodeMap[host]
	if selectedNode == nil {
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
