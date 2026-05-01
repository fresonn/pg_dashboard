package inmemory

import (
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type Store[V comparable] struct {
	ttlcache *ttlcache.Cache[string, V]
}

func New[V comparable](defaultTTL time.Duration) *Store[V] {

	cache := ttlcache.New(
		ttlcache.WithTTL[string, V](defaultTTL),
	)

	go cache.Start()

	return &Store[V]{
		ttlcache: cache,
	}
}

func (i *Store[V]) Set(key string, value V, ttl time.Duration) *ttlcache.Item[string, V] {
	return i.ttlcache.Set(key, value, ttl)
}

func (c *Store[V]) Get(key string, opts ...ttlcache.Option[string, V]) *ttlcache.Item[string, V] {
	return c.ttlcache.Get(key, opts...)
}

func (c *Store[V]) DeleteAll() {
	c.ttlcache.DeleteAll()
}
