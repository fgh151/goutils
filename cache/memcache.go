package cache

import (
	"encoding/json"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/rgglez/gormcache"
	"log"
	"os"
)

type MemCache struct {
	Client *memcache.Client
}

func GetMemClient() *memcache.Client {
	return memcache.New(os.Getenv("CACHE_SRV"))
}

func (m *MemCache) Get(key string) (interface{}, error) {
	v, err := m.Client.Get(key)
	if err != nil {
		log.Println("No found from cache key: " + key)
		return nil, err
	}

	var value interface{}
	err = json.Unmarshal(v.Value, &value)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (m *MemCache) Set(key string, value interface{}, ttl int32) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = m.Client.Set(&memcache.Item{Key: key, Value: data, Expiration: ttl})
	if err != nil {
		return err
	}
	return nil
}

func (m *MemCache) Delete(key string) error {
	err := m.Client.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (m *MemCache) Flush() error {
	err := m.Client.FlushAll()
	if err != nil {
		return err
	}
	return nil
}

func (m *MemCache) GormClient() gormcache.CacheClient {
	return gormcache.NewMemcacheClient(m.Client)
}
