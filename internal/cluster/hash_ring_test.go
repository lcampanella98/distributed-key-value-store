package cluster

import (
	"strconv"
	"testing"
)

func TestIsInRange(t *testing.T) {
	tests := []struct {
		name              string
		sortedRingHashes  []uint64
		replicationFactor int
		nodeHash          uint64
		expectedStart     uint64
		expectedEnd       uint64
	}{
		{
			name:              "3 nodes, 2 replicas, target node not in ring",
			sortedRingHashes:  []uint64{100, 1000, 10000},
			replicationFactor: 2,
			nodeHash:          5000,
			expectedStart:     100,
			expectedEnd:       5000,
		},
		{
			name:              "3 nodes, 2 replicas, target node not in ring",
			sortedRingHashes:  []uint64{100, 1000, 10000},
			replicationFactor: 2,
			nodeHash:          50,
			expectedStart:     1000,
			expectedEnd:       50,
		},
		{
			name:              "3 nodes, 3 replicas, target node not in ring",
			sortedRingHashes:  []uint64{100, 1000, 10000},
			replicationFactor: 3,
			nodeHash:          500,
			expectedStart:     1000,
			expectedEnd:       500,
		},
		{
			name:              "2 nodes, 3 replicas, target node not in ring",
			sortedRingHashes:  []uint64{100, 1000},
			replicationFactor: 3,
			nodeHash:          500,
			expectedStart:     500,
			expectedEnd:       500,
		},
		{
			name:              "3 nodes, 2 replicas, target node in ring",
			sortedRingHashes:  []uint64{100, 1000, 10000},
			replicationFactor: 2,
			nodeHash:          10000,
			expectedStart:     100,
			expectedEnd:       10000,
		},
		{
			name:              "3 nodes, 3 replicas, target node in ring",
			sortedRingHashes:  []uint64{100, 1000, 10000},
			replicationFactor: 3,
			nodeHash:          100,
			expectedStart:     100,
			expectedEnd:       100,
		},
		{
			name:              "1 node, 3 replicas, target node in ring",
			sortedRingHashes:  []uint64{100},
			replicationFactor: 3,
			nodeHash:          100,
			expectedStart:     100,
			expectedEnd:       100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sortedRing []nodeAndHash
			for i, hash := range tt.sortedRingHashes {
				node := Node{Name: strconv.Itoa(i)}
				sortedRing = append(sortedRing, nodeAndHash{
					Node: node,
					hash: hash,
				})
			}

			Ring.sortedRing = sortedRing
			Replicas = tt.replicationFactor
			start, end := Ring.GetStartAndEndRangeForNodeHash(tt.nodeHash)
			if start != tt.expectedStart || end != tt.expectedEnd {
				t.Errorf("start expected=%v actual=%v, end expected=%v actual=%v\n",
					tt.expectedStart, start, tt.expectedEnd, end)
			}
		})
	}
}
