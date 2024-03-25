package cache

import (
	"github.com/rgglez/gormcache"
	"os"
)

type AppCache interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl int32) error
	Delete(key string) error
	Flush() error
	GormClient() gormcache.CacheClient
}

func GetClient() AppCache {
	var client AppCache
	switch os.Getenv("CACHE_TYPE") {
	case "redis":
		client = &Redis{}
	case "memcache":
	default:
		client = &MemCache{Client: GetMemClient()}
	}

	return client
}
