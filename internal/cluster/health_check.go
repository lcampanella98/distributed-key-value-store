package cluster

import (
	"fmt"
	"time"

	"github.com/lcampanella98/distributed-key-value-store/internal/client"
	"github.com/lcampanella98/distributed-key-value-store/internal/mymetrics"
)

type NodeStatus struct {
	Node
	Alive    bool
	LastSeen time.Time
}

var nodeStatuses = make(map[string]*NodeStatus)

var durationBeforeDead time.Duration = time.Second * 8
var healthCheckInterval time.Duration = time.Second * 4

func StartHealthChecks(allNodes []Node) {
	var peerNodes []Node
	for _, node := range allNodes {
		nodeStatuses[node.Name] = &NodeStatus{
			Node:     node,
			Alive:    true,
			LastSeen: time.Now(),
		}
		// don't run health check on this node, only on other nodes. this node's nodeStatus.Alive will always be true
		if node.Name != ThisNode.Name {
			peerNodes = append(peerNodes, node)
		}
	}

	go func() {
		for {
			time.Sleep(healthCheckInterval)
			changed := healthCheckPeers(peerNodes)
			if changed {
				aliveNodes := GetAliveNodes()
				mymetrics.M.SetAliveNodes(int64(len(aliveNodes)))
				fmt.Println("Rebuilding hash ring...")
				rebuildHashRingWithNodes(aliveNodes)
			}
		}
	}()
}

type healthCheckResult struct {
	node Node
	ok   bool
}

func GetAliveNodes() []Node {
	var aliveNodes []Node
	for _, status := range nodeStatuses {
		if status.Alive {
			aliveNodes = append(aliveNodes, status.Node)
		}
	}
	return aliveNodes
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
