package cluster

import (
	"fmt"

	"github.com/lcampanella98/distributed-key-value-store/internal/cache"
	"github.com/lcampanella98/distributed-key-value-store/internal/client"
	"github.com/lcampanella98/distributed-key-value-store/internal/types"
)

type coordinatorConfig struct {
	writeMode string
}

var config coordinatorConfig

func SetCoordinatorConfig(writeMode string) {
	if writeMode != "strict" && writeMode != "best_effort" {
		panic("Invalid write mode " + writeMode)
	}
	config = coordinatorConfig{
		writeMode: writeMode,
	}
}

type putInReplicaResult struct {
	i   int
	res types.PutResponse
	err error
}

func CoordinatorPut(key, value string, replicaSet []Node, myIndexInReplicaSet int) (types.PutResponse, error) {
	var res types.PutResponse

	putInReplica := func(i int, node Node, results chan<- putInReplicaResult) {
		response, err := client.PutWithCoordinator(key, value, node.Addr, ThisNode.Name)
		results <- putInReplicaResult{i: i, res: response, err: err}
	}

	results := make(chan putInReplicaResult, len(replicaSet))
	numRoutines := 0
	for i, node := range replicaSet {
		if i == myIndexInReplicaSet {
			// coordinator node happens to be the owner/in replica set
			// no need make a separate network call to put, since this is the current node just update locally
			cache.Put(key, value)
			response := types.PutResponse{Ok: true, OnNode: ThisNode.Name, CacheSize: cache.Size()}
			if i == 0 {
				res = response
			}
		} else {
			go putInReplica(i, node, results)
			numRoutines++
		}
	}
	var errs []error
	for range numRoutines {
		result := <-results
		if result.err != nil {
			fmt.Printf("Error occurred in put to server %s: %v\n", replicaSet[result.i].Addr, result.err)
			errs = append(errs, result.err)
		}
		if result.i == 0 {
			res = result.res
		}
	}
	close(results)

	if len(errs) > 0 {
		return res, errs[0]
	} else {
		return res, nil
	}
}
