package cacher

import (
	"log"
	"sync"
)

var cache Cache
var mux sync.Mutex

// GetCache Get an instance of cache
func GetCache() Cache {
	mux.Lock()
	if cache == nil {
		rc := newRedisCache()
		err := rc.initialize()
		if err != nil {
			log.Fatal(err)
		}
		cache = rc
	}
	mux.Unlock()
	return cache
}
