package api

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

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
	ownerNode := cluster.Ring.GetOwnerNode(key)
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

func put(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	key := q.Get("key")
	value := q.Get("value")
	coordinator := q.Get("coordinator")
	isCoordinator := coordinator == ""
	// fmt.Println("Put " + key + "=" + value)

	var res types.PutResponse
	replicaSet := cluster.Ring.GetReplicaSet(key)

	myIndexInReplicaSet := slices.IndexFunc(replicaSet, func(node cluster.Node) bool {
		return node.Name == cluster.ThisNode.Name
	})

	if !isCoordinator && myIndexInReplicaSet != -1 {
		// this node is in replica set, and not the coordinator, so only put into this node's cache
		cache.Put(key, value)
		res = types.PutResponse{Ok: true, OnNode: cluster.ThisNode.Name, CacheSize: cache.Size()}
	} else if isCoordinator {
		response, err := cluster.CoordinatorPut(key, value, replicaSet, myIndexInReplicaSet)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res = response
	} else {
		http.Error(w, "This node is not coordinator and is not in replica set", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

// Clears the cache on this node. Useful for testing purposes.
func clearCache(w http.ResponseWriter, req *http.Request) {
	cache.Clear()
	w.WriteHeader(http.StatusOK)
}

func health(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// kills the current node immediately. useful for testing purposes
func kill(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Killing this node on client request")
	w.WriteHeader(http.StatusOK)
	go func() {
		time.Sleep(10 * time.Millisecond)
		os.Exit(0)
	}()
}

func GetHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/get", get)
	mux.HandleFunc("/put", put)
	mux.HandleFunc("/clear", clearCache)
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/kill", kill)
	return mux
}
