package cluster

import (
	"github.com/cespare/xxhash"
)

func hash(s string) uint64 {
	return xxhash.Sum64String(s)
}
