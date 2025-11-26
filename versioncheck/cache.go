package versioncheck

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	// CacheTTL is how long cached version info is considered valid
	CacheTTL = 48 * time.Hour
)

// CacheEntry represents a cached version check result
type CacheEntry struct {
	LatestVersion string    `json:"latest_version"`
	CheckedAt     time.Time `json:"checked_at"`
}

// IsExpired returns whether the cache entry has exceeded its TTL
func (c *CacheEntry) IsExpired() bool {
	return time.Since(c.CheckedAt) > CacheTTL
}

// loadCache reads and parses the cache file
// Returns nil if cache doesn't exist or is invalid
func loadCache() (*CacheEntry, error) {
	cachePath, err := getCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[DEBUG] No cache file found at %s", cachePath)
			return nil, nil
		}
		return nil, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		log.Printf("[DEBUG] Failed to parse cache file: %s", err)
		return nil, err
	}

	return &entry, nil
}

// saveCache writes the cache entry to disk
func saveCache(entry *CacheEntry) error {
	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return err
	}

	log.Printf("[DEBUG] Saved version check cache to %s", cachePath)
	return nil
}

// getCachePath returns the full path to the cache file using platform-specific cache directory
func getCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "tflint", "version_check_cache.json"), nil
}
