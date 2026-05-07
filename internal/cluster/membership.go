package cluster

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"sync"
)

type Node struct {
	Name string
	Addr string
}

type nodeAndHash struct {
	Node
	hash uint64
}

type HashRing struct {
	mu         sync.RWMutex
	sortedRing []nodeAndHash
}

func compare(a, b nodeAndHash) int {
	return cmp.Compare(a.hash, b.hash)
}

var ring = HashRing{}
var ThisNode Node
var Replicas int

func PrintHashRing() {
	ring.mu.RLock()
	defer ring.mu.RUnlock()
	fmt.Println("=== Hash Ring ===")
	for i, node := range ring.sortedRing {
		var prevNode nodeAndHash
		var diff uint64
		if i == 0 {
			prevNode = ring.sortedRing[len(ring.sortedRing)-1]
			diff = math.MaxUint64 - prevNode.hash + node.hash
		} else {
			prevNode = ring.sortedRing[i-1]
			diff = node.hash - prevNode.hash
		}
		perc := float64(diff) / float64(math.MaxUint64) * 100
		fmt.Printf("Node %s owns %.2f%% of hash space\n", node.Name, perc)
	}
}

func InitHashRing(nodes []Node, thisNode Node, replicas int) {
	ring.mu.Lock()
	defer ring.mu.Unlock()
	ThisNode = thisNode
	Replicas = replicas
	for _, node := range nodes {
		element := nodeAndHash{Node: node, hash: hash(node.Name)}
		ring.sortedRing = append(ring.sortedRing, element)
	}
	slices.SortFunc(ring.sortedRing, compare)
}

func rebuildHashRing(nodeStatuses map[string]NodeStatus) {
	newRing := make([]nodeAndHash, 0)
	for _, status := range nodeStatuses {
		if !status.Alive {
			continue
		}
		element := nodeAndHash{Node: status.Node, hash: hash(status.Node.Name)}
		newRing = append(newRing, element)
	}
	slices.SortFunc(newRing, compare)
	ring.mu.Lock()
	defer ring.mu.Unlock()
	ring.sortedRing = newRing
}

// BinarySearchCeil returns the index of the first element >= target.
// If target is greater than all elements, it wraps around and returns 0.
func (h *HashRing) binarySearchCeil(keyHash uint64) int {
	if len(h.sortedRing) == 0 {
		return -1 // handle empty array
	}

	lo, hi := 0, len(h.sortedRing)-1
	result := -1 // will hold the best candidate index

	for lo <= hi {
		mid := lo + (hi-lo)/2

		if h.sortedRing[mid].hash >= keyHash {
			result = mid // mid is a valid candidate (>= target)
			hi = mid - 1 // try to find an earlier one
		} else {
			lo = mid + 1 // arr[mid] < target, go right
		}
	}

	// If no element >= target was found, wrap around to index 0
	if result == -1 {
		return 0
	}
	return result
}

func GetOwnerNode(key string) Node {
	keyHash := hash(key)
	ring.mu.RLock()
	defer ring.mu.RUnlock()
	primaryIndex := ring.binarySearchCeil(keyHash)
	return ring.sortedRing[primaryIndex].Node
}

func GetReplicaSet(key string) []Node {
	replicaSet := make([]Node, 0)
	keyHash := hash(key)
	ring.mu.RLock()
	defer ring.mu.RUnlock()
	primaryIndex := ring.binarySearchCeil(keyHash)
	// calculate effective replicas. if too many nodes have gone dead, we can only have at most the number of alive nodes as the number of replicas
	effectiveReplicas := min(Replicas, len(ring.sortedRing))
	for i := range effectiveReplicas {
		replicaSet = append(replicaSet, ring.sortedRing[(primaryIndex+i)%len(ring.sortedRing)].Node)
	}
	return replicaSet
}
