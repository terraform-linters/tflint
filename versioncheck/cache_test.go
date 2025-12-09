package versioncheck

import (
	"testing"
	"time"
)

func TestCacheEntry_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		checkedAt time.Time
		want      bool
	}{
		{
			name:      "fresh cache (1 hour old)",
			checkedAt: time.Now().Add(-1 * time.Hour),
			want:      false,
		},
		{
			name:      "fresh cache (24 hours old)",
			checkedAt: time.Now().Add(-24 * time.Hour),
			want:      false,
		},
		{
			name:      "expired cache (49 hours old)",
			checkedAt: time.Now().Add(-49 * time.Hour),
			want:      true,
		},
		{
			name:      "just expired (48 hours + 1 minute)",
			checkedAt: time.Now().Add(-48*time.Hour - 1*time.Minute),
			want:      true,
		},
		{
			name:      "just fresh (47 hours)",
			checkedAt: time.Now().Add(-47 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &CacheEntry{
				LatestVersion: "0.60.0",
				CheckedAt:     tt.checkedAt,
			}

			got := entry.IsExpired()
			if got != tt.want {
				t.Errorf("CacheEntry.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
