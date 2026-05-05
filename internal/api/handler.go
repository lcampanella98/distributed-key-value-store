package api

import (
	"net/http"

	"encoding/json"

	"github.com/lcampanella98/distributed-key-value-store/internal/cache"
)

type getResponse struct {
	Value  string `json:"value"`
	Exists bool   `json:"exists"`
}

func get(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	key := q.Get("key")
	value, ok := cache.Cache[key]
	res := getResponse{Value: value, Exists: ok}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

type putResponse struct {
	Ok bool `json:"ok"`
}

func put(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	key := q.Get("key")
	value := q.Get("value")
	cache.Cache[key] = value
	res := putResponse{Ok: true}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

func InitHandler() {
	http.HandleFunc("/get", get)
	http.HandleFunc("/put", put)

	http.ListenAndServe(":8090", nil)

}
