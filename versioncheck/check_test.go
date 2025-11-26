package versioncheck

import (
	"os"
	"testing"

	"github.com/hashicorp/go-version"
)

func TestEnabled(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{
			name:     "not set - enabled by default",
			envValue: "",
			want:     true,
		},
		{
			name:     "disabled",
			envValue: "1",
			want:     false,
		},
		{
			name:     "invalid value - enabled by default",
			envValue: "invalid",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original env
			original := os.Getenv("TFLINT_DISABLE_VERSION_CHECK")
			defer func() {
				if original == "" {
					os.Unsetenv("TFLINT_DISABLE_VERSION_CHECK")
				} else {
					os.Setenv("TFLINT_DISABLE_VERSION_CHECK", original)
				}
			}()

			if tt.envValue == "" {
				os.Unsetenv("TFLINT_DISABLE_VERSION_CHECK")
			} else {
				os.Setenv("TFLINT_DISABLE_VERSION_CHECK", tt.envValue)
			}

			got := Enabled()
			if got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name          string
		current       string
		latest        string
		wantAvailable bool
		wantError     bool
	}{
		{
			name:          "update available",
			current:       "0.59.0",
			latest:        "0.60.0",
			wantAvailable: true,
			wantError:     false,
		},
		{
			name:          "v prefix stripped",
			current:       "0.59.0",
			latest:        "v0.60.0",
			wantAvailable: true,
			wantError:     false,
		},
		{
			name:          "already latest",
			current:       "0.60.0",
			latest:        "0.60.0",
			wantAvailable: false,
			wantError:     false,
		},
		{
			name:          "invalid latest version",
			current:       "0.60.0",
			latest:        "invalid",
			wantAvailable: false,
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := version.Must(version.NewVersion(tt.current))

			got, err := compareVersions(current, tt.latest)
			if (err != nil) != tt.wantError {
				t.Errorf("compareVersions() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if got.Available != tt.wantAvailable {
					t.Errorf("compareVersions() Available = %v, want %v", got.Available, tt.wantAvailable)
				}

				expectedLatest := tt.latest
				if expectedLatest[0] == 'v' {
					expectedLatest = expectedLatest[1:]
				}
				if got.Latest != expectedLatest {
					t.Errorf("compareVersions() Latest = %v, want %v", got.Latest, expectedLatest)
				}
			}
		})
	}
}
