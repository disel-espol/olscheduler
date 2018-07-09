package balancer

import (
	"errors"
	"hash/fnv"

	"../schutil"
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

	return &workers[targetIndex], nil
}
