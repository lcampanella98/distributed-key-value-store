package cluster

import (
	"fmt"

	"github.com/lcampanella98/distributed-key-value-store/internal/cache"
	"github.com/lcampanella98/distributed-key-value-store/internal/client"
	"github.com/lcampanella98/distributed-key-value-store/internal/types"
)

func pull(fromNode Node, thisNode Node, results chan<- types.RepairResponse) {
	res, err := client.Repair(thisNode.Name, fromNode.Addr)
	// if there was an err the data in the response will be an empty map which is fine
	if err != nil {
		fmt.Printf("Error pulling from node %v (%v data items)\n", fromNode.Name, len(res.Data))
	} else {
		fmt.Printf("Pulled %v data items from node %v\n", len(res.Data), fromNode.Name)
	}
	results <- res
}

func PullFromPeers(nodes []Node, thisNode Node) {
	go func() {
		results := make(chan types.RepairResponse, len(nodes))
		nRequests := 0
		for _, node := range nodes {
			if node.Name == thisNode.Name {
				continue
			}
			go pull(node, thisNode, results)
			nRequests++
		}
		for range nRequests {
			result := <-results
			cache.PutAll(result.Data)
		}
	}()
}
