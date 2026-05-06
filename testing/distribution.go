package main

import (
	"fmt"

	"github.com/lcampanella98/distributed-key-value-store/internal/client"
)

func TestDistribution() {
	fmt.Println("Testing Distribution of KV Store")
	addr := "http://localhost:8080"
	cacheSizes := make(map[string]int)

	for range 1000 {
		key := generateRandomString(10)
		value := generateRandomString(10)
		client.Put(key, value, addr)
		getRes, _ := client.Get(key, addr)
		if getRes.Value != value {
			fmt.Println("retrieved value not equal to expected value")
		}
		cacheSizes[getRes.OnNode] = getRes.CacheSize
	}
	fmt.Println(cacheSizes)

}
