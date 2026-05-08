package cache

import (
	"sync"

	"github.com/lcampanella98/distributed-key-value-store/internal/hashing"
)

var cache = make(map[string]string)
var mut = sync.RWMutex{}

func Get(key string) (string, bool) {
	mut.RLock()
	defer mut.RUnlock()
	v, ok := cache[key]
	return v, ok
}

func Put(key, value string) {
	mut.Lock()
	defer mut.Unlock()
	cache[key] = value
}

func Clear() {
	mut.Lock()
	defer mut.Unlock()
	clear(cache)
}

func Size() int {
	mut.RLock()
	defer mut.RUnlock()
	return len(cache)
}

func GetAllInHashRange(rangeStart, rangeEnd uint64) map[string]string {
	data := make(map[string]string)
	mut.RLock()
	defer mut.RUnlock()
	for k, v := range cache {
		h := hashing.Hash(k)
		if hashing.IsInRange(rangeStart, rangeEnd, h) {
			data[k] = v
		}
	}
	return data
}

func PutAll(m map[string]string) {
	mut.Lock()
	defer mut.Unlock()
	for k, v := range m {
		cache[k] = v
	}
}
