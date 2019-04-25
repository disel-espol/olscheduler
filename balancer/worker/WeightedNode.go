package worker

import "net/url"

type WeightedNode struct {
	url    url.URL
	weight int
}

func (n WeightedNode) GetWeight() int {
	return n.weight
}

func (n WeightedNode) GetURL() url.URL {
	return n.url
}

func NewWeightedNode(workerUrl url.URL, weight int) WeightedNode {
	return WeightedNode{workerUrl, weight}
}
