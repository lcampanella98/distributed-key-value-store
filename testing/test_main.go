package main

import (
	"fmt"
	"time"

	"github.com/lcampanella98/distributed-key-value-store/internal/client"
)

func main() {
	client.Init(false)

	start := time.Now()
	TestErrorStatistics()
	elapsed := time.Since(start)
	fmt.Printf("Execution took %s\n", elapsed)
}
