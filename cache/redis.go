package cache

import (
	"github.com/rgglez/gormcache"
)

type Redis struct{}

func GetRedisClient() {}

func (r *Redis) Get(key string) (interface{}, error) {
	return nil, nil
}

func (r *Redis) Set(key string, value interface{}, ttl int32) error {
	return nil
}

func (r *Redis) Delete(key string) error {
	return nil
}

func (r *Redis) Flush() error {
	return nil
}

func (r *Redis) GormClient() gormcache.CacheClient {
	return gormcache.NewMemcacheClient(nil)
}
