package service

import (
	"context"
	"time"

	"log/slog"

	"github.com/sol-armada/website/internal/cache"
)

// CachedAdminService wraps AdminService with caching
type CachedAdminService struct {
	*AdminService
	cache  *cache.RedisCache
	logger *slog.Logger
	ttl    time.Duration
}

// NewCachedAdminService creates a cached admin service
func NewCachedAdminService(adminService *AdminService, redisCache *cache.RedisCache, logger *slog.Logger) *CachedAdminService {
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
		cas.logger.Debug("Cache hit for overview stats", "key", key)
		return stats, nil
	}

	// Cache miss, get from database
	stats, err := cas.AdminService.GetOverviewStats(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := cas.cache.Set(ctx, key, stats, cas.ttl); err != nil {
		cas.logger.Warn("Failed to cache stats", "error", err)
	}

	return stats, nil
}

func (cas *CachedAdminService) GetTokenLedgerAnalytics(ctx context.Context) (*TokenLedgerAnalytics, error) {
	key := cache.CacheKey{Prefix: cache.KeyAdminTokenAnalytics}.String()

	var stats *TokenLedgerAnalytics
	if err := cas.cache.GetJSON(ctx, key, &stats); err == nil {
		cas.logger.Debug("Cache hit for token ledger analytics", "key", key)
		return stats, nil
	}

	stats, err := cas.AdminService.GetTokenLedgerAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	if err := cas.cache.Set(ctx, key, stats, cas.ttl); err != nil {
		cas.logger.Warn("Failed to cache token ledger analytics", "error", err)
	}

	return stats, nil
}

func (cas *CachedAdminService) GetAttendanceAnalytics(ctx context.Context) (*AttendanceAnalytics, error) {
	key := cache.CacheKey{Prefix: cache.KeyAdminAttendanceAnalytics}.String()

	var stats *AttendanceAnalytics
	if err := cas.cache.GetJSON(ctx, key, &stats); err == nil {
		cas.logger.Debug("Cache hit for attendance analytics", "key", key)
		return stats, nil
	}

	stats, err := cas.AdminService.GetAttendanceAnalytics(ctx)
	if err != nil {
		return nil, err
	}

	if err := cas.cache.Set(ctx, key, stats, cas.ttl); err != nil {
		cas.logger.Warn("Failed to cache attendance analytics", "error", err)
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
		cache.CacheKey{Prefix: cache.KeyAdminTokenAnalytics}.String(),
		cache.CacheKey{Prefix: cache.KeyAdminAttendanceAnalytics}.String(),
	}
	return cas.cache.Del(ctx, keys...)
}
