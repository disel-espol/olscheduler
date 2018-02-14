package balancer

import (
	"../schutil"
	"hash/fnv"
	"errors"
	"log"
)

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

func SelectWorkerPkgAware(workers []schutil.Worker, 
							pkgs []string, 
							threshold int) (*schutil.Worker, error) {
	if len(workers) == 0 {
		return nil, errors.New("Can't select worker, Workers empty")
	}

	if len(pkgs) == 0 {
		return nil, errors.New("Can't select worker, No largest package, pkgs empty")
	}

	for i, _ := range workers {
		log.Printf("--> worker: %s with load: %d", workers[i].URL.String(), workers[i].Load)
	}

	largestPkg := pkgs[0]
	targetIndex1 := h1(largestPkg)%uint32(len(workers))
	targetIndex2 := h2(largestPkg)%uint32(len(workers))

	targetIndex := targetIndex2
	if workers[targetIndex1].Load < workers[targetIndex2].Load {
		targetIndex = targetIndex1
	}
	
	if workers[targetIndex].Load > threshold { // Find least loaded
		targetIndex = 0
		for i := 1; i < len(workers); i++ {
			if workers[i].Load < workers[targetIndex].Load {
				targetIndex = uint32(i)
			}
		}
	}

	return &workers[targetIndex], nil
}