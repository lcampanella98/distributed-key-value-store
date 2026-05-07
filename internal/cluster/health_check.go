package cluster

import (
	"fmt"
	"time"

	"github.com/lcampanella98/distributed-key-value-store/internal/client"
)

type NodeStatus struct {
	Node
	Alive    bool
	LastSeen time.Time
}

var nodeStatuses = make(map[string]*NodeStatus)

var durationBeforeDead time.Duration = time.Second * 8

func StartHealthChecks(allNodes []Node) {
	for _, node := range allNodes {
		nodeStatuses[node.Name] = &NodeStatus{
			Node:     node,
			Alive:    true,
			LastSeen: time.Now(),
		}
	}

	go func() {
		peerNodes := make([]Node, 0)
		for _, node := range allNodes {
			if node.Name != ThisNode.Name {
				peerNodes = append(peerNodes, node)
			}
		}
		for {
			changed := healthCheckPeers(peerNodes)
			if changed {
				fmt.Println("Rebuilding hash ring...")
				rebuildHashRing()
			}
			time.Sleep(4 * time.Second)
		}
	}()
}

type healthCheckResult struct {
	node Node
	ok   bool
}

func healthCheckPeers(peers []Node) bool {
	changed := false
	healthCheck := func(node Node, results chan<- healthCheckResult) {
		err := client.Health(node.Addr)
		results <- healthCheckResult{node: node, ok: err == nil}
	}
	results := make(chan healthCheckResult, len(peers))
	for _, node := range peers {
		go healthCheck(node, results)
	}
	curTime := time.Now()
	for range peers {
		result := <-results

		status := nodeStatuses[result.node.Name]
		if result.ok {
			if !status.Alive {
				fmt.Printf("Node %s is now alive\n", result.node.Name)
				changed = true
			}
			status.Alive = true
			status.LastSeen = curTime
		} else {
			if status.Alive {
				// an alive node failed health check. determine whether to mark it as dead
				elapsedSinceHealthy := curTime.Sub(status.LastSeen)
				if elapsedSinceHealthy > durationBeforeDead {
					fmt.Printf("Node %s has gone dead\n", result.node.Name)
					status.Alive = false
					changed = true
				}
			} else {
				// a dead node failed another health check. no need to do anything here
			}
		}

	}
	return changed
}
