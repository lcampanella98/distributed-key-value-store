package main

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/lcampanella98/distributed-key-value-store/internal/client"
)

func TestConstantTraffic() {
	fmt.Println("Testing with Constant Traffic")
	writeMode := "best_effort"
	addr := []string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082"}
	coordinatorIdx := 0
	killIdx := 2
	putThenGetAfter := 10 * time.Second

	// set address of coordinator
	address := addr[coordinatorIdx]

	for _, address := range addr {
		client.Clear(address)
	}
	putCounterHealthy := 0
	putErrCounterHealthy := 0
	getCounterHealthy := 0
	getErrCounterHealthy := 0
	getWrongValueCounterHealthy := 0

	putCounterPreRecovery := 0
	putErrCounterPreRecovery := 0
	getCounterPreRecovery := 0
	getErrCounterPreRecovery := 0
	getWrongValueCounterPreRecovery := 0

	putCounterPostRecovery := 0
	putErrCounterPostRecovery := 0
	getCounterPostRecovery := 0
	getErrCounterPostRecovery := 0
	getWrongValueCounterPostRecovery := 0

	var curPutCounter *int = &putCounterHealthy
	var curPutErrCounter *int = &putErrCounterHealthy
	var curGetCounter *int = &getCounterHealthy
	var curGetErrCounter *int = &getErrCounterHealthy
	var curGetWrongValueCounter *int = &getWrongValueCounterHealthy

	staleReadCount := 0
	retryDelays := map[string]time.Duration{
		"100ms":     100 * time.Millisecond,
		"1 second":  1 * time.Second,
		"2 seconds": 2 * time.Second,
		"6 seconds": 6 * time.Second,
	}
	staleReadAfterRetry := make(map[string]int)
	for delayStr := range retryDelays {
		staleReadAfterRetry[delayStr] = 0
	}
	firstStaleRead := time.Time{}
	lastStaleRead := time.Time{}

	mu := sync.Mutex{}

	doGet := func(address string, key string, expectedValue string, isRetry bool) bool {
		getRes, getErr := client.Get(key, address)
		mu.Lock()
		(*curGetCounter)++
		mu.Unlock()
		if getErr != nil {
			mu.Lock()
			(*curGetErrCounter)++
			mu.Unlock()
			// fmt.Printf("Get error: %v\n", putErr)
		}
		if getErr == nil && getRes.Value != expectedValue {
			if !isRetry {
				mu.Lock()
				(*curGetWrongValueCounter)++
				if staleReadCount == 0 {
					firstStaleRead = time.Now()
				}
				lastStaleRead = time.Now()
				staleReadCount++
				mu.Unlock()
			}
			// fmt.Println("wrong get value")
			return true
		}
		return false
		// if getRes.Value != value {
		// 	fmt.Println("retrieved value not equal to expected value")
		// }
	}

	doPutRandomPairThenGet := func(address string) {
		key := generateRandomString(20)
		value := generateRandomString(20)
		putRes, putErr := client.Put(key, value, address)
		(*curPutCounter)++
		if putErr != nil {
			fmt.Printf("Put error: %v\n", putErr)
		}
		if !putRes.Ok {
			fmt.Println("Put not OK")
		}
		if putErr != nil || !putRes.Ok {
			mu.Lock()
			(*curPutErrCounter)++
			mu.Unlock()
		}
		go func() {
			var staleRead bool
			time.Sleep(putThenGetAfter)
			staleRead = doGet(address, key, value, false)
			if !staleRead {
				return
			}

			for delayStr, delayDur := range retryDelays {
				go func(str string, dur time.Duration) {
					time.Sleep(dur)
					staleRead = doGet(address, key, value, true)
					if !staleRead {
						return
					} else {
						staleReadAfterRetry[str]++
					}
				}(delayStr, delayDur)
			}
		}()
	}

	// build server
	cmd := exec.Command("go", "build", "-o", `C:\Users\enzoc\repos\distributed-key-value-store\main.exe`, `C:\Users\enzoc\repos\distributed-key-value-store\cmd\node\main.go`)
	err := cmd.Run()
	if err != nil {
		panic("Failed to build server executable")
	}

	// stage 1: healthy cluster for 5 seconds
	for range 5000 {
		doPutRandomPairThenGet(address)
		time.Sleep(time.Millisecond)
	}

	// stage 2: kill node
	client.Kill(addr[killIdx])

	curPutCounter = &putCounterPreRecovery
	curPutErrCounter = &putErrCounterPreRecovery
	curGetCounter = &getCounterPreRecovery
	curGetErrCounter = &getErrCounterPreRecovery
	curGetWrongValueCounter = &getWrongValueCounterPreRecovery

	for range 17000 {
		doPutRandomPairThenGet(address)
		time.Sleep(time.Millisecond)
	}

	// stage 3: revive node
	ctx, cancel := context.WithCancel(context.Background())

	cmd = exec.CommandContext(
		ctx,
		`C:\Users\enzoc\repos\distributed-key-value-store\main.exe`,
		"-port=8082",
		"-nodes=localhost:8080,localhost:8081,localhost:8082",
		"-replicas=3",
		"-writeMode="+writeMode,
	)

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	curPutCounter = &putCounterPostRecovery
	curPutErrCounter = &putErrCounterPostRecovery
	curGetCounter = &getCounterPostRecovery
	curGetErrCounter = &getErrCounterPostRecovery
	curGetWrongValueCounter = &getWrongValueCounterPostRecovery

	nodeRecoveredTime := time.Now()

	for range 10000 {
		doPutRandomPairThenGet(address)
		time.Sleep(time.Millisecond)
	}

	// wait for gets to finish
	time.Sleep(putThenGetAfter + (6 * time.Second))

	fmt.Println("- === HEALTHY CLUSTER ===")
	fmt.Printf("- %.2f%% of puts errored (%v)\n", float64(putErrCounterHealthy)/float64(putCounterHealthy)*100, putErrCounterHealthy)
	fmt.Printf("- %.2f%% of gets errored (%v)\n", float64(getErrCounterHealthy)/float64(getCounterHealthy)*100, getErrCounterHealthy)
	fmt.Printf("- %.2f%% of gets returned incorrect value (%v)\n", float64(getWrongValueCounterHealthy)/float64(getCounterHealthy)*100, getWrongValueCounterHealthy)

	fmt.Println("- === 1 NODE DEAD CLUSTER ===")
	fmt.Printf("- %.2f%% of puts errored (%v)\n", float64(putErrCounterPreRecovery)/float64(putCounterPreRecovery)*100, putErrCounterPreRecovery)
	fmt.Printf("- %.2f%% of gets errored (%v)\n", float64(getErrCounterPreRecovery)/float64(getCounterPreRecovery)*100, getErrCounterPreRecovery)
	fmt.Printf("- %.2f%% of gets returned incorrect value (%v)\n", float64(getWrongValueCounterPreRecovery)/float64(getCounterPreRecovery)*100, getWrongValueCounterPreRecovery)

	fmt.Println("- === NODE RECOVERED CLUSTER ===")
	fmt.Printf("- %.2f%% of puts errored (%v)\n", float64(putErrCounterPostRecovery)/float64(putCounterPostRecovery)*100, putErrCounterPostRecovery)
	fmt.Printf("- %.2f%% of gets errored (%v)\n", float64(getErrCounterPostRecovery)/float64(getCounterPostRecovery)*100, getErrCounterPostRecovery)
	fmt.Printf("- %.2f%% of gets returned incorrect value (%v)\n", float64(getWrongValueCounterPostRecovery)/float64(getCounterPostRecovery)*100, getWrongValueCounterPostRecovery)

	fmt.Println("- === Stale Reads ===")
	if staleReadCount > 0 {
		for durationStr, staleReadCounter := range staleReadAfterRetry {
			fmt.Printf("- %.2f%% of stale reads remained after %v\n", float64(staleReadCounter)/float64(staleReadCount)*100, durationStr)
		}
	}

	fmt.Printf("- Node recovered at: %v\n", nodeRecoveredTime)
	fmt.Printf("- First stale read: %v and stale read window was %v\n", firstStaleRead, lastStaleRead.Sub(firstStaleRead))

	cancel()

	err = cmd.Wait()
	fmt.Println("child exited:", err)
	fmt.Println("finished")

}
