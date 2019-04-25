package worker

import "net/url"

type Node struct {
	url  url.URL
	Load uint
}

func (n *Node) GetURL() url.URL {
	return n.url
}

func NewNode(url url.URL) *Node {
	return &Node{url, 0}
}
