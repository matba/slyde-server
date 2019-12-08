package cacher

// Cache The interface that represents a generic cache
type Cache interface {
	// Add a key and value with a timout to cache
	AddKeyValue(key string, value string, timeout int) error
	// gets the value for key or nil if it is not present
	GetKeyValue(key string) (string, error)
}
