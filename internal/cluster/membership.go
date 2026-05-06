package cluster

import (
	"cmp"
	"fmt"
	"math"
	"slices"
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
	sortedRing []*nodeAndHash
}

func compare(a, b *nodeAndHash) int {
	return cmp.Compare(a.hash, b.hash)
}

var ring = HashRing{[]*nodeAndHash{}}
var ThisNode Node
var Replicas int

func PrintHashRing() {
	for i, node := range ring.sortedRing {
		var prevNode *nodeAndHash
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
	ThisNode = thisNode
	Replicas = replicas
	for _, node := range nodes {
		element := nodeAndHash{Node: node, hash: hash(node.Name)}
		fmt.Printf("computed hash %d\n", element.hash)
		ring.sortedRing = append(ring.sortedRing, &element)
	}
	slices.SortFunc(ring.sortedRing, compare)
}

// BinarySearchCeil returns the index of the first element >= target.
// If target is greater than all elements, it wraps around and returns 0.
func binarySearchCeil(arr []*nodeAndHash, target uint64) int {
	if len(arr) == 0 {
		return -1 // handle empty array
	}

	lo, hi := 0, len(arr)-1
	result := -1 // will hold the best candidate index

	for lo <= hi {
		mid := lo + (hi-lo)/2

		if arr[mid].hash >= target {
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
	nodeIndex := binarySearchCeil(ring.sortedRing, keyHash)
	return ring.sortedRing[nodeIndex].Node
}

func GetReplicaSet(key string) []Node {
	replicaSet := make([]Node, 0)
	keyHash := hash(key)
	nodeIndex := binarySearchCeil(ring.sortedRing, keyHash)
	for i := range Replicas {
		replicaSet = append(replicaSet, ring.sortedRing[(nodeIndex+i)%len(ring.sortedRing)].Node)
	}
	return replicaSet
}
