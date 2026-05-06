package main

import (
	"fmt"

	"github.com/lcampanella98/distributed-key-value-store/internal/client"
)

func TestDistribution() {
	fmt.Println("Testing Distribution of KV Store")
	addr := []string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082"}
	addrIdx := 0
	cacheSizes := make(map[string]int)

	for _, address := range addr {
		client.Clear(address)
	}

	for range 1000 {
		key := generateRandomString(50)
		value := generateRandomString(50)
		client.Put(key, value, addr[addrIdx])

		getRes, _ := client.Get(key, addr[addrIdx])
		if getRes.Value != value {
			fmt.Println("retrieved value not equal to expected value")
		}
		cacheSizes[getRes.OnNode] = getRes.CacheSize
		addrIdx = (addrIdx + 1) % len(addr)
	}
	fmt.Println(cacheSizes)
	total := 0
	for _, v := range cacheSizes {
		total += v
	}
	fmt.Printf("%d\n", total)
}
