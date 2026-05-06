package main

import (
	"fmt"

	"github.com/lcampanella98/distributed-key-value-store/internal/client"
)

func TestErrorStatistics() {
	fmt.Println("Testing with error statistics")
	addr := []string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082"}
	addrIdx := 0
	totalIterations := 1000
	getErrCount := 0
	putErrCount := 0
	wrongGetValueCount := 0

	for _, address := range addr {
		client.Clear(address)
	}

	doPutRandomPair := func(address string) {
		key := generateRandomString(20)
		value := generateRandomString(20)
		putRes, putErr := client.Put(key, value, address)
		if putErr != nil {
			fmt.Printf("Put error: %v\n", putErr)
		}
		if !putRes.Ok {
			fmt.Println("Put not OK")
		}
		if putErr != nil || !putRes.Ok {
			putErrCount++
		}

		getRes, getErr := client.Get(key, address)
		if getErr != nil {
			getErrCount++
			fmt.Printf("Get error: %v\n", putErr)
		}
		if getRes.Value != value {
			wrongGetValueCount++
			fmt.Println("wrong get value")
		}
		// if getRes.Value != value {
		// 	fmt.Println("retrieved value not equal to expected value")
		// }
	}

	for range totalIterations {
		address := addr[addrIdx]
		doPutRandomPair(address)

		// addrIdx = (addrIdx + 1) % len(addr)
	}
	fmt.Printf("%.2f%% of puts errored\n", float64(putErrCount)/float64(totalIterations)*100)
	fmt.Printf("%.2f%% of gets errored\n", float64(getErrCount)/float64(totalIterations)*100)
	fmt.Printf("%.2f%% of gets returned incorrect value\n", float64(wrongGetValueCount)/float64(totalIterations)*100)

	fmt.Println("finished")
}
