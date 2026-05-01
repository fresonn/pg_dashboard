package cache

import (
	"context"
	"dashboard/api/internal/config"
	"dashboard/api/internal/infra/inmemory"
	"dashboard/api/internal/infra/logger"
	"dashboard/api/internal/model/cluster"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type Cache struct {
	logger logger.Logger
	cache  *inmemory.Store[any]
}

func New(config config.AppConfig, logger logger.Logger) *Cache {

	cache := inmemory.New[any](5 * time.Minute)

	return &Cache{
		logger: logger,
		cache:  cache,
	}
}

func (c *Cache) Clear(ctx context.Context) {
	c.logger.DebugContext(ctx, "clear cluster cache")
	c.cache.DeleteAll()
}

func (c *Cache) SetPgVersion(ctx context.Context, version cluster.PostgresVersion) {
	c.logger.DebugContext(ctx, "cache set", "key", pgVersionKey, "value", version)
	c.cache.Set(pgVersionKey, version, ttlcache.NoTTL)
}

func (c *Cache) PgVersion(ctx context.Context) (cluster.PostgresVersion, bool) {

	item := c.cache.Get(pgVersionKey)
	if item == nil {
		c.logger.DebugContext(ctx, "cache miss", "key", pgVersionKey)
		return cluster.PostgresVersion{}, false
	}

	c.logger.DebugContext(ctx, "cache hit", "key", item.Key(), "expires_at", item.ExpiresAt())

	val, ok := item.Value().(cluster.PostgresVersion)
	if !ok {
		c.logger.ErrorContext(ctx, "cache type cast failed", "key", item.Key())
		return cluster.PostgresVersion{}, false
	}

	return val, true
}

func (c *Cache) SetClusterUptime(ctx context.Context, version cluster.PostgresUptime) {
	c.logger.DebugContext(ctx, "cache set", "key", clusterUptimeKey, "value", version)
	c.cache.Set(clusterUptimeKey, version, ttlcache.NoTTL)
}

func (c *Cache) ClusterUptime(ctx context.Context) (cluster.PostgresUptime, bool) {

	item := c.cache.Get(clusterUptimeKey)
	if item == nil {
		c.logger.DebugContext(ctx, "cache miss", "key", clusterUptimeKey)
		return cluster.PostgresUptime{}, false
	}

	c.logger.DebugContext(ctx, "cache hit", "key", item.Key(), "expires_at", item.ExpiresAt())

	val, ok := item.Value().(cluster.PostgresUptime)
	if !ok {
		c.logger.ErrorContext(ctx, "cache type cast failed", "key", item.Key())
		return cluster.PostgresUptime{}, false
	}

	return val, true
}
