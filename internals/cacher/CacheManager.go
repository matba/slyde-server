package cacher

import (
	"sync"
)

var cache Cache
var mux sync.Mutex

// GetCache Get an instance of cache
func GetCache() Cache {
	mux.Lock()
	if cache == nil {
		rc := newRedisCache()
		cache = rc
	}
	mux.Unlock()
	return cache
}
