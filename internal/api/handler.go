package api

import (
	"fmt"
	"net/http"

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
		res, err = client.Get(key, ownerNode.Node.Addr)
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
	// fmt.Println("Put " + key + "=" + value)

	var res types.PutResponse
	ownerNode := cluster.GetOwnerNode(key)
	if ownerNode.Name == cluster.ThisNode.Name {
		// fmt.Printf("this node owns key %s\n", key)
		cache.Cache[key] = value
		res = types.PutResponse{Ok: true, OnNode: ownerNode.Name, CacheSize: len(cache.Cache)}
	} else {
		// fmt.Printf("node %s owns key %s\n", ownerNode.Name, key)
		var err error
		res, err = client.Put(key, value, ownerNode.Node.Addr)
		if err != nil {
			fmt.Printf("Error occurred in put: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

func GetHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/get", get)
	mux.HandleFunc("/put", put)
	return mux
}
