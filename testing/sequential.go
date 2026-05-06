package main

import (
	"fmt"

	"github.com/lcampanella98/distributed-key-value-store/internal/client"
)

func TestSequential() {
	fmt.Println("Testing Distribution of KV Store")
	addr := []string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082"}
	addrIdx := 0
	cacheSizes := make(map[string]int)

	for _, address := range addr {
		client.Clear(address)
	}

	for range 1000 {
		key := generateRandomString(20)
		value := generateRandomString(20)
		putRes, putErr := client.Put(key, value, addr[addrIdx])
		if putErr != nil {
			fmt.Printf("Put error: %v\n", putErr)
		}
		if !putRes.Ok {
			fmt.Println("Put not OK")
		}

		getRes, getErr := client.Get(key, addr[addrIdx])
		if getErr != nil {
			fmt.Printf("Get error: %v\n", putErr)
		}
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
