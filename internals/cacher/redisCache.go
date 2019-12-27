package cacher

import (
	"github.com/gomodule/redigo/redis"
)

type redisCache struct {
	cache redis.Conn
}

func newRedisCache() *redisCache {
	c := redisCache{}
	return &c
}

func (c *redisCache) initialize() error {
	// Initialize the redis connection to a redis instance running on your local machine
	conn, err := redis.DialURL("redis://localhost")
	if err != nil {
		return err
	}
	// Assign the connection to the package level `cache` variable
	c.cache = conn
	return nil
}

func (c *redisCache) AddKeyValue(key string, value string, timeout int) error {
	_, err := c.cache.Do("SETEX", key, timeout, value)
	return err
}

func (c *redisCache) GetKeyValue(key string) (string, error) {
	response, err := c.cache.Do("GET", key)
	if err != nil {
		return "", err
	}

	if response == nil {
		return "", NotFound
	}

	return string(response.([]byte)), err
}

func (c *redisCache) DeleteKey(key string) error {
	_, err := c.cache.Do("DEL", key)
	return err
}
