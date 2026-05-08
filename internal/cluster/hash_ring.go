package cluster

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"sync"

	"github.com/lcampanella98/distributed-key-value-store/internal/hashing"
)

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

var Ring = HashRing{}

func (ring *HashRing) PrintHashRing() {
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

// BinarySearchCeil returns the index of the first element >= target.
// If target is greater than all elements, it wraps around and returns 0.
func (h *HashRing) binarySearchCeil(keyHash uint64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
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

func (ring *HashRing) GetOwnerNode(key string) Node {
	keyHash := hashing.Hash(key)
	ring.mu.RLock()
	defer ring.mu.RUnlock()
	primaryIndex := ring.binarySearchCeil(keyHash)
	return ring.sortedRing[primaryIndex].Node
}

func (ring *HashRing) GetReplicaSet(key string) []Node {
	replicaSet := make([]Node, 0)
	keyHash := hashing.Hash(key)
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

func mod(a, b int) int {
	return (a%b + b) % b
}

// TODO handle case when node is already in hash ring!
func (ring *HashRing) GetStartAndEndRangeForNode(node Node) (uint64, uint64) {
	hash := hashing.Hash(node.Name)
	return ring.GetStartAndEndRangeForNodeHash(hash)
}

func (ring *HashRing) GetStartAndEndRangeForNodeHash(nodeHash uint64) (uint64, uint64) {
	ring.mu.RLock()
	defer ring.mu.RUnlock()
	effectiveReplicas := min(Replicas, len(ring.sortedRing))
	nextNodeIdx := ring.binarySearchCeil(nodeHash)
	startNodeIdx := mod(nextNodeIdx-effectiveReplicas, len(ring.sortedRing))
	rangeStart, rangeEnd := ring.sortedRing[startNodeIdx].hash, nodeHash
	return rangeStart, rangeEnd
}

func (ring *HashRing) rebuildHashRing(nodes []Node) {
	var newRing []nodeAndHash
	for _, node := range nodes {
		element := nodeAndHash{Node: node, hash: hashing.Hash(node.Name)}
		newRing = append(newRing, element)
	}
	slices.SortFunc(newRing, compare)
	ring.mu.Lock()
	defer ring.mu.Unlock()
	ring.sortedRing = newRing
}
