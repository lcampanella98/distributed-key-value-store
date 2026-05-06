package cache

import "sync"

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
