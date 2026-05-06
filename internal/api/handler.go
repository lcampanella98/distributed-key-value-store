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
		value, ok := cache.Cache[key]
		res = types.GetResponse{Value: value, Exists: ok, OnNode: ownerNode.Name, CacheSize: len(cache.Cache)}
	} else {
		// fmt.Printf("node %s owns key %s\n", ownerNode.Name, key)
		var err error
		res, err = client.Get(key, ownerNode.Addr)
		if err != nil {
			fmt.Printf("Error occurred in get: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
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

	if !isCoordinator && thisNodeIndex != -1 {
		// this node is owner or replica, and not the coordinator, so only put into this node's cache
		cache.Cache[key] = value
		res = types.PutResponse{Ok: true, OnNode: cluster.ThisNode.Name, CacheSize: len(cache.Cache)}
	} else if isCoordinator {
		for i, node := range ownerAndReplicas {
			var response types.PutResponse
			var err error
			if i == thisNodeIndex {
				// coordinator node happens to be the owner/in replica set
				// no need make a separate network call to put, since this is the current node just update locally
				cache.Cache[key] = value
				response = types.PutResponse{Ok: true, OnNode: cluster.ThisNode.Name, CacheSize: len(cache.Cache)}
			} else {
				response, err = client.PutWithCoordinator(key, value, node.Addr, cluster.ThisNode.Name)
				if err != nil {
					fmt.Printf("Error occurred in put to server %s: %v\n", node.Addr, err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			if i == 0 {
				// return the response of the owner node. this populates the OnNode and CacheSize fields with owner's values
				res = response
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

func clearCache(w http.ResponseWriter, req *http.Request) {
	clear(cache.Cache)
	w.WriteHeader(http.StatusOK)
}

func GetHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/get", get)
	mux.HandleFunc("/put", put)
	mux.HandleFunc("/clear", clearCache)
	return mux
}
