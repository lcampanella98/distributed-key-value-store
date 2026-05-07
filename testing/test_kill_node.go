package main

import (
	"fmt"
	"time"

	"github.com/lcampanella98/distributed-key-value-store/internal/client"
)

func TestKillNode() {
	fmt.Println("Testing killing a node")
	addr := []string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082"}
	coordinatorIdx := 0
	killIdx := 2

	keysPerStep := 1000
	waitForFailureDetection := 12 * time.Second

	// set address of coordinator
	address := addr[coordinatorIdx]

	for _, address := range addr {
		client.Clear(address)
	}

	doGet := func(address string, key string, expectedValue string, errCounter *int, wrongValueCounter *int) {
		getRes, getErr := client.Get(key, address)
		if getErr != nil {
			(*errCounter)++
			// fmt.Printf("Get error: %v\n", putErr)
		}
		if getRes.Value != expectedValue {
			(*wrongValueCounter)++
			// fmt.Println("wrong get value")
		}
		// if getRes.Value != value {
		// 	fmt.Println("retrieved value not equal to expected value")
		// }
	}

	doPutRandomPair := func(address string, store *map[string]string, errCounter *int) {
		key := generateRandomString(20)
		value := generateRandomString(20)
		(*store)[key] = value
		putRes, putErr := client.Put(key, value, address)
		if putErr != nil {
			fmt.Printf("Put error: %v\n", putErr)
		}
		if !putRes.Ok {
			fmt.Println("Put not OK")
		}
		if putErr != nil || !putRes.Ok {
			(*errCounter)++
		}
	}

	// step 1: put random pairs into healthy cluster
	m1 := make(map[string]string)
	putErrsHealthy, getErrsHealthy, getWrongValueHealthy := 0, 0, 0

	for range keysPerStep {
		doPutRandomPair(address, &m1, &putErrsHealthy)
	}
	for k, v := range m1 {
		doGet(address, k, v, &getErrsHealthy, &getWrongValueHealthy)
	}

	// step 2: kill a node (that's not the coordinator)
	client.Kill(addr[killIdx])
	time.Sleep(100 * time.Millisecond)

	putErrsDegraded, getErrsDegraded, getWrongValueDegraded := 0, 0, 0
	// step 2a: read and compare M1
	for k, v := range m1 {
		doGet(address, k, v, &getErrsDegraded, &getWrongValueDegraded)
	}
	// step 2b: put new random values
	m2 := make(map[string]string)
	for range keysPerStep {
		doPutRandomPair(address, &m2, &putErrsDegraded)
	}

	// step 3: recovered cluster
	// step 3a: wait for failure detection to kick in
	time.Sleep(waitForFailureDetection)

	// step 3b: read from M1
	getErrsRecoveredM1, getWrongValueRecoveredM1 := 0, 0
	for k, v := range m1 {
		doGet(address, k, v, &getErrsRecoveredM1, &getWrongValueRecoveredM1)
	}
	getErrsRecoveredM2, getWrongValueRecoveredM2 := 0, 0
	for k, v := range m2 {
		doGet(address, k, v, &getErrsRecoveredM2, &getWrongValueRecoveredM2)
	}
	putErrsRecovered := 0
	m3 := make(map[string]string) // unused
	for range keysPerStep {
		doPutRandomPair(address, &m3, &putErrsRecovered)
	}
	fmt.Println("=== HEALTHY CLUSTER ===")
	fmt.Printf("%.2f%% of puts errored\n", float64(putErrsHealthy)/float64(keysPerStep)*100)
	fmt.Printf("%.2f%% of gets from M1 errored\n", float64(getErrsHealthy)/float64(keysPerStep)*100)
	fmt.Printf("%.2f%% of gets from M1 returned incorrect value\n", float64(getWrongValueHealthy)/float64(keysPerStep)*100)

	fmt.Println("=== DEGRADED CLUSTER (pre-failure detection) ===")
	fmt.Printf("%.2f%% of gets from M1 errored\n", float64(getErrsDegraded)/float64(keysPerStep)*100)
	fmt.Printf("%.2f%% of gets from M1 returned incorrect value\n", float64(getWrongValueDegraded)/float64(keysPerStep)*100)
	fmt.Printf("%.2f%% of puts errored\n", float64(putErrsDegraded)/float64(keysPerStep)*100)

	fmt.Println("=== RECOVERED CLUSTER (post-failure detection) ===")
	fmt.Printf("%.2f%% of gets from M1 errored\n", float64(getErrsRecoveredM1)/float64(keysPerStep)*100)
	fmt.Printf("%.2f%% of gets from M1 returned incorrect value\n", float64(getWrongValueRecoveredM1)/float64(keysPerStep)*100)
	fmt.Printf("%.2f%% of gets from M2 errored\n", float64(getErrsRecoveredM2)/float64(keysPerStep)*100)
	fmt.Printf("%.2f%% of gets from M2 returned incorrect value\n", float64(getWrongValueRecoveredM2)/float64(keysPerStep)*100)
	fmt.Printf("%.2f%% of puts errored\n", float64(putErrsRecovered)/float64(keysPerStep)*100)

	fmt.Println("finished")
}
