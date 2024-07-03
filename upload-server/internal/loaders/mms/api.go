package mms

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/redis/go-redis/v9"
)

type APIConfigLoader struct {
	BaseURL     string
	HTTPClient  *http.Client
	RedisClient *redis.Client
	CacheTTL    time.Duration
}

// NewAPIConfigLoader initializes the APIConfigLoader with HTTP and Redis clients.
func NewAPIConfigLoader(baseURL string, redisAddr string, cacheTTL time.Duration) *APIConfigLoader {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &APIConfigLoader{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		RedisClient: rdb,
		CacheTTL:    cacheTTL,
	}
}

// LoadConfig retrieves the configuration from the cache or API.
func (l *APIConfigLoader) LoadConfig(ctx context.Context, path string) ([]byte, error) {
	cacheKey := "config:" + path

	// Attempt to retrieve from cache
	cachedConfig, err := l.RedisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		return cachedConfig, nil
	}
	if err != redis.Nil {
		return nil, err
	}

	// Cache miss: retrieve from API
	url, err := url.JoinPath(l.BaseURL, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := l.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, errors.New(models.MMS_CONFIG_NOTFOUND_MSG)
		}
		return nil, errors.New(models.MMS_CONFIG_RET_FAILED_MSG)
	}

	configData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Store in cache
	err = l.RedisClient.Set(ctx, cacheKey, configData, l.CacheTTL).Err()
	if err != nil {
		return nil, err
	}

	return configData, nil
}

// InvalidateCache invalidates the cache for a specific path.
func (l *APIConfigLoader) InvalidateCache(ctx context.Context, path string) error {
	cacheKey := "config:" + path
	return l.RedisClient.Del(ctx, cacheKey).Err()
}
