package hashing

import (
	"github.com/cespare/xxhash"
)

func Hash(s string) uint64 {
	return xxhash.Sum64String(s)
}
