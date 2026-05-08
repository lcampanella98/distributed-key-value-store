package main

import (
	"fmt"

	"github.com/lcampanella98/distributed-key-value-store/internal/client"
)

func TestHighConcurrency() {
	fmt.Println("Testing High Concurrency")
	addr := []string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082"}
	addrIdx := 0

	for _, address := range addr {
		client.Clear(address)
	}

	doPutRandomValue := func(key string, address string) {
		value := generateRandomString(20)
		putRes, putErr := client.Put(key, value, address)
		if putErr != nil {
			fmt.Printf("Put error: %v\n", putErr)
		}
		if !putRes.Ok {
			fmt.Println("Put not OK")
		}

		_, getErr := client.Get(key, address)
		if getErr != nil {
			fmt.Printf("Get error: %v\n", putErr)
		}
		// if getRes.Value != value {
		// 	fmt.Println("retrieved value not equal to expected value")
		// }
	}
	key := generateRandomString(20)

	for range 4000 {
		address := addr[addrIdx]
		go doPutRandomValue(key, address)

		addrIdx = (addrIdx + 1) % len(addr)
	}
	fmt.Println("finished")
}
