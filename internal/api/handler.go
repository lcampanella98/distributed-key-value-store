package api

import (
	"fmt"
	"net/http"
	"slices"

	"encoding/json"

	"github.com/lcampanella98/distributed-key-value-store/internal/cache"
	"github.com/lcampanella98/distributed-key-value-store/internal/client"
	"github.com/lcampanella98/distributed-key-value-store/internal/cluster"
	"github.com/lcampanella98/distributed-key-value-store/internal/types"
)

func get(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	key := q.Get("key")
	// fmt.Println("Get " + key)
	var res types.GetResponse
	ownerNode := cluster.GetOwnerNode(key)
	if ownerNode.Name == cluster.ThisNode.Name {
		// fmt.Printf("this node owns key %s\n", key)
		value, ok := cache.Get(key)
		res = types.GetResponse{Value: value, Exists: ok, OnNode: ownerNode.Name, CacheSize: cache.Size()}
	} else {
		// fmt.Printf("node %s owns key %s\n", ownerNode.Name, key)
		var err error
		res, err = client.Get(key, ownerNode.Addr)
		if err != nil {
			fmt.Printf("Error occurred in get: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

type putInReplicaResult struct {
	i   int
	res types.PutResponse
	err error
}

func put(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	key := q.Get("key")
	value := q.Get("value")
	coordinator := q.Get("coordinator")
	isCoordinator := coordinator == ""
	// fmt.Println("Put " + key + "=" + value)

	var res types.PutResponse
	ownerAndReplicas := cluster.GetReplicaSet(key)
	thisNodeIndex := slices.IndexFunc(ownerAndReplicas, func(node cluster.Node) bool {
		return node.Name == cluster.ThisNode.Name
	})

	putInReplica := func(i int, node cluster.Node, results chan<- putInReplicaResult) {
		response, err := client.PutWithCoordinator(key, value, node.Addr, cluster.ThisNode.Name)
		results <- putInReplicaResult{i: i, res: response, err: err}
	}

	if !isCoordinator && thisNodeIndex != -1 {
		// this node is owner or replica, and not the coordinator, so only put into this node's cache
		cache.Put(key, value)
		res = types.PutResponse{Ok: true, OnNode: cluster.ThisNode.Name, CacheSize: cache.Size()}
	} else if isCoordinator {
		results := make(chan putInReplicaResult, len(ownerAndReplicas))
		numRoutines := 0
		for i, node := range ownerAndReplicas {
			if i == thisNodeIndex {
				// coordinator node happens to be the owner/in replica set
				// no need make a separate network call to put, since this is the current node just update locally
				cache.Put(key, value)
				response := types.PutResponse{Ok: true, OnNode: cluster.ThisNode.Name, CacheSize: cache.Size()}
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
				fmt.Printf("Error occurred in put to server %s: %v\n", ownerAndReplicas[result.i].Addr, result.err)
				errs = append(errs, result.err)
			}
			if result.i == 0 {
				res = result.res
			}
		}
		close(results)

		if len(errs) > 0 {
			// send the first error to client
			http.Error(w, errs[0].Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

func clearCache(w http.ResponseWriter, req *http.Request) {
	cache.Clear()
	w.WriteHeader(http.StatusOK)
}

func GetHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/get", get)
	mux.HandleFunc("/put", put)
	mux.HandleFunc("/clear", clearCache)
	return mux
}
