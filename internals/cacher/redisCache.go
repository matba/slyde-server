package cacher

import (
	"github.com/gomodule/redigo/redis"
)

type redisCache struct {
}

func newRedisCache() *redisCache {
	c := redisCache{}
	return &c
}

func (c *redisCache) createConnection() (redis.Conn, error) {
	// Initialize the redis connection to a redis instance running on your local machine
	conn, err := redis.DialURL("redis://localhost")
	if err != nil {
		return nil, err
	}
	// Assign the connection to the package level `cache` variable
	return conn, nil
}

func (c *redisCache) AddKeyValue(key string, value string, timeout int) error {
	con, err := c.createConnection()
	if err != nil {
		return err
	}
	_, err = con.Do("SETEX", key, timeout, value)
	con.Close()
	return err
}

func (c *redisCache) GetKeyValue(key string) (string, error) {
	con, err := c.createConnection()
	if err != nil {
		return "", err
	}
	response, err := con.Do("GET", key)
	con.Close()
	if err != nil {
		return "", err
	}

	if response == nil {
		return "", NotFound
	}

	return string(response.([]byte)), err
}

func (c *redisCache) DeleteKey(key string) error {
	con, err := c.createConnection()
	if err != nil {
		return err
	}
	_, err = con.Do("DEL", key)
	con.Close()
	return err
}
