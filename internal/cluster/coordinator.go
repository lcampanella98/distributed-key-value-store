package cluster

import (
	"errors"
	"fmt"
	"time"

	"github.com/lcampanella98/distributed-key-value-store/internal/cache"
	"github.com/lcampanella98/distributed-key-value-store/internal/client"
	"github.com/lcampanella98/distributed-key-value-store/internal/mymetrics"
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
		startTime := time.Now()

		response, err := client.PutWithCoordinator(key, value, node.Addr, ThisNode.Name)

		latency := time.Since(startTime)
		mymetrics.M.IncReplicationRequestsTotal()
		mymetrics.M.ObserveReplicationLatency(latency)
		if err != nil {
			mymetrics.M.IncReplicationFailed()
		}

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
	var primaryError error = nil
	var errs []error
	for range numRoutines {
		result := <-results
		if result.err != nil {
			fmt.Printf("Error occurred in put to server %s: %v\n", replicaSet[result.i].Addr, result.err)
			errs = append(errs, result.err)
			if result.i == 0 {
				primaryError = result.err
			}
		}
		if result.i == 0 {
			res = result.res
		}
	}
	close(results)

	switch config.writeMode {
	case "strict":
		if len(errs) > 0 {
			return res, errs[0]
		} else {
			return res, nil
		}
	case "best_effort":
		if primaryError != nil {
			return res, primaryError
		} else {
			return res, nil
		}
	default:
		panic("Couldn't match write mode " + config.writeMode)
	}

}

type getFromReplicaResult struct {
	i   int
	res types.GetResponse
	err error
}

func CoordinatorGet(key string, replicaSet []Node, myIndexInReplicaSet int) (types.GetResponse, error) {
	var res types.GetResponse

	getFromReplica := func(i int, node Node, results chan<- getFromReplicaResult) {
		startTime := time.Now()

		response, err := client.GetWithCoordinator(key, node.Addr, ThisNode.Name)

		latency := time.Since(startTime)
		mymetrics.M.IncGetFromReplicaRequestsTotal()
		mymetrics.M.ObserveGetFromReplicaLatency(latency)
		if err != nil {
			mymetrics.M.IncGetFromReplicaFailed()
		}

		results <- getFromReplicaResult{i: i, res: response, err: err}
	}

	results := make(chan getFromReplicaResult, len(replicaSet))
	// this flags whether at least 1 replica responded and has data for this key
	foundKeyExists := false
	// this flags whether at least 1 replica responded and has no data for this key
	foundKeyDoesntExist := false
	var keyDoesntExistRes types.GetResponse

	for i, node := range replicaSet {
		if i == myIndexInReplicaSet {
			// coordinator node happens to be the owner/in replica set
			// no need make a separate network call to put, since this is the current node just update locally
			value, exists := cache.Get(key)
			response := types.GetResponse{Exists: exists, Value: value, OnNode: ThisNode.Name, CacheSize: cache.Size()}
			if !exists {
				if !foundKeyDoesntExist {
					foundKeyDoesntExist = true
					keyDoesntExistRes = response
				}
			} else {
				res = response
				foundKeyExists = true
				break
			}
		} else {
			getFromReplica(i, node, results)
			result := <-results
			if result.err == nil {
				if !result.res.Exists {
					if !foundKeyDoesntExist {
						foundKeyDoesntExist = true
						keyDoesntExistRes = result.res
					}
				} else {
					res = result.res
					foundKeyExists = true
					break
				}
			}
		}
	}
	close(results)
	var err error = nil
	if !foundKeyExists {
		// no replica responded with data for this key. but a replica still might have responded with no data for the key rather than an error
		if foundKeyDoesntExist {
			// at least 1 replica responded that it has no data for this key
			res = keyDoesntExistRes
		} else {
			// all replicas failed
			err = errors.New("All replicas errored")
		}
	}
	return res, err
}
