package getproviders

import (
	"github.com/apparentlymart/go-versions/versions"
	"github.com/apparentlymart/go-versions/versions/constraints"

	"github.com/terraform-linters/tflint/terraform/addrs"
)

// Version represents a particular single version of a provider.
type Version = versions.Version

// UnspecifiedVersion is the zero value of Version, representing the absense
// of a version number.
var UnspecifiedVersion Version = versions.Unspecified

// VersionConstraints represents a set of version constraints, which can
// define the membership of a VersionSet by exclusion.
type VersionConstraints = constraints.IntersectionSpec

// Requirements gathers together requirements for many different providers
// into a single data structure, as a convenient way to represent the full
// set of requirements for a particular configuration or state or both.
//
// If an entry in a Requirements has a zero-length VersionConstraints then
// that indicates that the provider is required but that any version is
// acceptable. That's different than a provider being absent from the map
// altogether, which means that it is not required at all.
type Requirements map[addrs.Provider]VersionConstraints

// Merge takes the requirements in the receiever and the requirements in the
// other given value and produces a new set of requirements that combines
// all of the requirements of both.
//
// The resulting requirements will permit only selections that both of the
// source requirements would've allowed.
func (r Requirements) Merge(other Requirements) Requirements {
	ret := make(Requirements)
	for addr, constraints := range r {
		ret[addr] = constraints
	}
	for addr, constraints := range other {
		ret[addr] = append(ret[addr], constraints...)
	}
	return ret
}

// ParseVersion parses a "semver"-style version string into a Version value,
// which is the version syntax we use for provider versions.
func ParseVersion(str string) (Version, error) {
	return versions.ParseVersion(str)
}

// ParseVersionConstraints parses a "Ruby-like" version constraint string
// into a VersionConstraints value.
func ParseVersionConstraints(str string) (VersionConstraints, error) {
	return constraints.ParseRubyStyleMulti(str)
}
