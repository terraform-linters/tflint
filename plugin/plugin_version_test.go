package plugin

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
)

func TestCheckSDKVersionSatisfiesConstraints(t *testing.T) {
	tests := []struct {
		name         string
		pluginName   string
		sdkVersion   string
		isJSONConfig bool
		wantErr      bool
		errContains  string
	}{
		// HCL config tests
		{
			name:         "HCL config with SDK 0.14.0 (below minimum)",
			pluginName:   "test",
			sdkVersion:   "0.14.0",
			isJSONConfig: false,
			wantErr:      true,
			errContains:  "incompatible",
		},
		{
			name:         "HCL config with SDK 0.16.0 (minimum)",
			pluginName:   "test",
			sdkVersion:   "0.16.0",
			isJSONConfig: false,
			wantErr:      false,
		},
		{
			name:         "HCL config with SDK 0.22.0 (above minimum)",
			pluginName:   "test",
			sdkVersion:   "0.22.0",
			isJSONConfig: false,
			wantErr:      false,
		},
		{
			name:         "HCL config with SDK 0.23.0",
			pluginName:   "test",
			sdkVersion:   "0.23.0",
			isJSONConfig: false,
			wantErr:      false,
		},

		// JSON config tests
		{
			name:         "JSON config with SDK 0.14.0 (way below minimum)",
			pluginName:   "test",
			sdkVersion:   "0.14.0",
			isJSONConfig: true,
			wantErr:      true,
			errContains:  "incompatible with JSON configuration",
		},
		{
			name:         "JSON config with SDK 0.16.0 (below JSON minimum)",
			pluginName:   "test",
			sdkVersion:   "0.16.0",
			isJSONConfig: true,
			wantErr:      true,
			errContains:  "incompatible with JSON configuration",
		},
		{
			name:         "JSON config with SDK 0.22.0 (below JSON minimum)",
			pluginName:   "test",
			sdkVersion:   "0.22.0",
			isJSONConfig: true,
			wantErr:      true,
			errContains:  "incompatible with JSON configuration",
		},
		{
			name:         "JSON config with SDK 0.23.0 (JSON minimum)",
			pluginName:   "test",
			sdkVersion:   "0.23.0",
			isJSONConfig: true,
			wantErr:      false,
		},
		{
			name:         "JSON config with SDK 0.24.0 (above JSON minimum)",
			pluginName:   "test",
			sdkVersion:   "0.24.0",
			isJSONConfig: true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := version.NewVersion(tt.sdkVersion)
			if err != nil {
				t.Fatalf("Failed to parse version %s: %v", tt.sdkVersion, err)
			}

			err = CheckSDKVersionSatisfiesConstraints(tt.pluginName, v, tt.isJSONConfig)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CheckSDKVersionSatisfiesConstraints() expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("CheckSDKVersionSatisfiesConstraints() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("CheckSDKVersionSatisfiesConstraints() unexpected error = %v", err)
				}
			}
		})
	}
}
