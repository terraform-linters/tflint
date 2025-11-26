package versioncheck

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
)

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	Available bool
	Latest    string
}

// Enabled returns whether version checking is enabled
func Enabled() bool {
	val := os.Getenv("TFLINT_DISABLE_VERSION_CHECK")
	if val == "" {
		return true
	}

	disabled, err := strconv.ParseBool(val)
	if err != nil {
		return true
	}

	return !disabled
}

// CheckForUpdate checks if a new version of tflint is available
// It returns UpdateInfo indicating if an update is available and the latest version string
// Errors are logged but not returned - failures should not break the version command
func CheckForUpdate(ctx context.Context, current *version.Version) (*UpdateInfo, error) {

	// Try to load from cache first
	cache, err := loadCache()
	if err != nil {
		log.Printf("[DEBUG] Failed to load version check cache: %s", err)
	} else if cache != nil && !cache.IsExpired() {
		log.Printf("[DEBUG] Using cached version check result")
		return compareVersions(current, cache.LatestVersion)
	}

	// Cache miss or expired, fetch from GitHub
	log.Printf("[DEBUG] Checking for TFLint updates...")
	latestVersion, err := fetchLatestRelease(ctx)
	if err != nil {
		return nil, err
	}

	// Save to cache (non-blocking, errors logged only)
	if err := saveCache(&CacheEntry{
		LatestVersion: latestVersion,
		CheckedAt:     time.Now(),
	}); err != nil {
		log.Printf("[DEBUG] Failed to save version check cache: %s", err)
	}

	return compareVersions(current, latestVersion)
}

// compareVersions compares current and latest versions and returns UpdateInfo
func compareVersions(current *version.Version, latestStr string) (*UpdateInfo, error) {
	// Strip leading "v" if present
	latestStr = strings.TrimPrefix(latestStr, "v")

	latest, err := version.NewVersion(latestStr)
	if err != nil {
		log.Printf("[DEBUG] Failed to parse latest version %q: %s", latestStr, err)
		return nil, err
	}

	log.Printf("[DEBUG] Current version: %s, Latest version: %s", current, latest)

	return &UpdateInfo{
		Available: latest.GreaterThan(current),
		Latest:    latestStr,
	}, nil
}
