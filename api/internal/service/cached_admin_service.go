package service

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sol-armada/website/internal/cache"
)

// CachedAdminService wraps AdminService with caching
type CachedAdminService struct {
	*AdminService
	cache  *cache.RedisCache
	logger *logrus.Logger
	ttl    time.Duration
}

// NewCachedAdminService creates a cached admin service
func NewCachedAdminService(adminService *AdminService, redisCache *cache.RedisCache, logger *logrus.Logger) *CachedAdminService {
	return &CachedAdminService{
		AdminService: adminService,
		cache:        redisCache,
		logger:       logger,
		ttl:          5 * time.Minute, // Cache TTL
	}
}

// GetOverviewStats retrieves stats with caching
func (cas *CachedAdminService) GetOverviewStats(ctx context.Context) (*AdminOverviewStats, error) {
	key := cache.CacheKey{Prefix: cache.KeyAdminOverview}.String()

	// Try cache first
	var stats *AdminOverviewStats
	if err := cas.cache.GetJSON(ctx, key, &stats); err == nil {
		cas.logger.WithField("key", key).Debug("Cache hit for overview stats")
		return stats, nil
	}

	// Cache miss, get from database
	stats, err := cas.AdminService.GetOverviewStats(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := cas.cache.Set(ctx, key, stats, cas.ttl); err != nil {
		cas.logger.WithError(err).Warnf("Failed to cache stats")
	}

	return stats, nil
}

// InvalidateOverviewCache clears the overview cache
func (cas *CachedAdminService) InvalidateOverviewCache(ctx context.Context) error {
	key := cache.CacheKey{Prefix: cache.KeyAdminOverview}.String()
	return cas.cache.Del(ctx, key)
}

// InvalidateAllCaches clears all admin caches
func (cas *CachedAdminService) InvalidateAllCaches(ctx context.Context) error {
	keys := []string{
		cache.CacheKey{Prefix: cache.KeyAdminOverview}.String(),
		cache.CacheKey{Prefix: cache.KeyAttendanceList}.String(),
	}
	return cas.cache.Del(ctx, keys...)
}
