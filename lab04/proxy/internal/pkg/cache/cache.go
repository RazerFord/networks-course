package cache

import (
	"encoding/json"
	"time"

	"github.com/peterbourgon/diskv/v3"
)

const (
	cacheSizeMax = 1024 * 1024 * 10 // 10MB
)

type Data struct {
	Key          string     `json:"key"`
	LastModified *time.Time `json:"last-modified"`
	CacheControl string     `json:"cache-control"`
	Value        []byte     `json:"value"`
}

func NewData(key string, lastModified *time.Time, cacheControl string, value []byte) *Data {
	return &Data{
		Key:          key,
		LastModified: lastModified,
		CacheControl: cacheControl,
		Value:        value,
	}
}

type Cache struct {
	d *diskv.Diskv
}

func NewCache() *Cache {
	c := Cache{
		d: diskv.New(
			diskv.Options{
				BasePath:     "cache",
				CacheSizeMax: cacheSizeMax,
			},
		),
	}
	return &c
}

func (c *Cache) Get(key string) (*Data, error) {
	bs, err := c.d.Read(key)

	if err != nil {
		return nil, err
	}

	data := new(Data)
	err = json.Unmarshal(bs, data)

	return data, err
}

func (c *Cache) Set(key string, data *Data) error {
	bs, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return c.d.Write(key, bs)
}

func (c *Cache) Erase(key string) error {
	return c.d.Erase(key)
}
