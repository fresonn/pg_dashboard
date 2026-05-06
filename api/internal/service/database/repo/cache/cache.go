package cache

import (
	"context"
	"dashboard/api/internal/config"
	"dashboard/api/internal/infra/inmemory"
	"dashboard/api/internal/infra/logger"
	"dashboard/api/internal/model/database"
	"strconv"
	"time"
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
	c.logger.DebugContext(ctx, "clear database cache")
	c.cache.DeleteAll()
}

func (c *Cache) SetDatabase(ctx context.Context, id int, db database.Database) {

	key := databaseByIdKey + strconv.Itoa(id)

	c.logger.DebugContext(ctx, "cache set", "key", key, "value", db)
	c.cache.Set(key, db, 30*time.Second)
}

func (c *Cache) Database(ctx context.Context, id int) (database.Database, bool) {

	key := databaseByIdKey + strconv.Itoa(id)

	item := c.cache.Get(key)
	if item == nil {
		c.logger.DebugContext(ctx, "cache miss", "key", key)
		return database.Database{}, false
	}

	c.logger.DebugContext(ctx, "cache hit", "key", item.Key(), "expires_at", item.ExpiresAt())

	val, ok := item.Value().(database.Database)
	if !ok {
		c.logger.ErrorContext(ctx, "cache type cast failed", "key", item.Key())
		return database.Database{}, false
	}

	return val, true
}
