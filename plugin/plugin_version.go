package plugin

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/terraform-linters/tflint/tflint"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Version constraints for plugin compatibility
var (
	// DefaultSDKVersionConstraints is the minimum SDK version for basic functionality
	DefaultSDKVersionConstraints = version.MustConstraints(version.NewConstraint(">= 0.16.0"))

	// JSONConfigSDKVersionConstraints is the minimum SDK version for JSON configuration support
	JSONConfigSDKVersionConstraints = version.MustConstraints(version.NewConstraint(">= 0.23.0"))

	// EphemeralMarksMinVersion is when ephemeral marks support was added
	EphemeralMarksMinVersion = version.Must(version.NewVersion("0.22.0"))
)

// CheckTFLintVersionConstraints validates if TFLint version meets the plugin's requirements
func CheckTFLintVersionConstraints(pluginName string, constraints version.Constraints) error {
	if !constraints.Check(tflint.Version) {
		return fmt.Errorf("Failed to satisfy version constraints; tflint-ruleset-%s requires %s, but TFLint version is %s", pluginName, constraints, tflint.Version)
	}
	return nil
}

// CheckSDKVersionSatisfiesConstraints validates if a plugin's SDK version meets the minimum requirements.
// For HCL configs, requires SDK >= 0.16.0. For JSON configs, requires SDK >= 0.23.0.
func CheckSDKVersionSatisfiesConstraints(pluginName string, sdkVersion *version.Version, isJSONConfig bool) error {
	// If sdkVersion is nil, the plugin doesn't have SDKVersion endpoint (SDK < 0.14)
	if sdkVersion == nil {
		return fmt.Errorf(`Plugin "%s" SDK version is incompatible. Compatible versions: %s`, pluginName, DefaultSDKVersionConstraints)
	}

	constraints := DefaultSDKVersionConstraints
	if isJSONConfig {
		constraints = JSONConfigSDKVersionConstraints
	}

	if !constraints.Check(sdkVersion) {
		if isJSONConfig {
			return fmt.Errorf(`Plugin "%s" SDK version (%s) is incompatible with JSON configuration. Minimum required: %s`, pluginName, sdkVersion, JSONConfigSDKVersionConstraints)
		}
		return fmt.Errorf(`Plugin "%s" SDK version (%s) is incompatible. Compatible versions: %s`, pluginName, sdkVersion, DefaultSDKVersionConstraints)
	}
	return nil
}

// SupportsEphemeralMarks checks if the plugin SDK version supports ephemeral marks
func SupportsEphemeralMarks(sdkVersion *version.Version) bool {
	if sdkVersion == nil {
		return false
	}
	return sdkVersion.GreaterThanOrEqual(EphemeralMarksMinVersion)
}

// IsSDKVersionUnimplemented checks if an error indicates the SDK version endpoint is not implemented
func IsSDKVersionUnimplemented(err error) bool {
	if st, ok := status.FromError(err); ok {
		return st.Code() == codes.Unimplemented
	}
	return false
}

// IsVersionConstraintsUnimplemented checks if an error indicates the version constraints endpoint is not implemented
func IsVersionConstraintsUnimplemented(err error) bool {
	if st, ok := status.FromError(err); ok {
		return st.Code() == codes.Unimplemented
	}
	return false
}

// ValidatePluginVersions checks plugin SDK version requirements and returns SDK versions for later use
// Note: TFLint version constraints are checked separately in launchPlugins before ApplyGlobalConfig
func ValidatePluginVersions(rulesetPlugin *Plugin, isJSONConfig bool) (map[string]*version.Version, error) {
	sdkVersions := map[string]*version.Version{}

	for name, ruleset := range rulesetPlugin.RuleSets {
		// Get SDK version
		sdkVersion, err := ruleset.SDKVersion()
		if err != nil {
			if IsSDKVersionUnimplemented(err) {
				// SDKVersion endpoint is available in tflint-plugin-sdk v0.14+.
				// Plugin is too old, treat as nil
				sdkVersion = nil
			} else {
				return nil, fmt.Errorf(`Failed to get plugin "%s" SDK version; %w`, name, err)
			}
		}

		// Check if SDK version meets minimum requirements
		if err := CheckSDKVersionSatisfiesConstraints(name, sdkVersion, isJSONConfig); err != nil {
			return nil, err
		}

		sdkVersions[name] = sdkVersion
	}

	return sdkVersions, nil
}
