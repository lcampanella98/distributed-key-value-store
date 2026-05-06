package benchmarks

import (
	"fmt"
	"time"

	"github.com/lcampanella98/distributed-key-value-store/internal/cache"
)

func printAllBenchmarks() {
	fmt.Printf("Cache keys: %v\n", len(cache.Cache))
}

func StartBenchmarks() {
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for range ticker.C {
			printAllBenchmarks()
		}
	}()
}
