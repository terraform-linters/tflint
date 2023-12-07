// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package addrs

import (
	"path"
	"strings"
)

// ModuleSource is the general type for all three of the possible module source
// address types. The concrete implementations of this are ModuleSourceLocal
// and ModuleSourceRemote.
type ModuleSource interface {
	// String returns a full representation of the address, including any
	// additional components that are typically implied by omission in
	// user-written addresses.
	//
	// We typically use this longer representation in error message, in case
	// the inclusion of normally-omitted components is helpful in debugging
	// unexpected behavior.
	String() string

	moduleSource()
}

var _ ModuleSource = ModuleSourceLocal("")
var _ ModuleSource = ModuleSourceRemote("")

var moduleSourceLocalPrefixes = []string{
	"./",
	"../",
	".\\",
	"..\\",
}

// ParseModuleSource parses a module source address as given in the "source"
// argument inside a "module" block in the configuration.
//
// Unlike Terraform, this function only categorizes sources into "local" and "remote".
func ParseModuleSource(raw string) (ModuleSource, error) {
	if isModuleSourceLocal(raw) {
		localAddr, err := parseModuleSourceLocal(raw)
		if err != nil {
			// This is to make sure we really return a nil ModuleSource in
			// this case, rather than an interface containing the zero
			// value of ModuleSourceLocal.
			return nil, err
		}
		return localAddr, nil
	}

	// Return all non-local sources assuming they are remote source.
	// Note that this is essentially useless for determining anything more
	// than "non-local".
	return ModuleSourceRemote(raw), nil
}

// ModuleSourceLocal is a ModuleSource representing a local path reference
// from the caller's directory to the callee's directory within the same
// module package.
//
// A "module package" here means a set of modules distributed together in
// the same archive, repository, or similar. That's a significant distinction
// because we always download and cache entire module packages at once,
// and then create relative references within the same directory in order
// to ensure all modules in the package are looking at a consistent filesystem
// layout. We also assume that modules within a package are maintained together,
// which means that cross-cutting maintenence across all of them would be
// possible.
//
// The actual value of a ModuleSourceLocal is a normalized relative path using
// forward slashes, even on operating systems that have other conventions,
// because we're representing traversal within the logical filesystem
// represented by the containing package, not actually within the physical
// filesystem we unpacked the package into. We should typically not construct
// ModuleSourceLocal values directly, except in tests where we can ensure
// the value meets our assumptions. Use ParseModuleSource instead if the
// input string is not hard-coded in the program.
type ModuleSourceLocal string

func parseModuleSourceLocal(raw string) (ModuleSourceLocal, error) {
	// As long as we have a suitable prefix (detected by ParseModuleSource)
	// there is no failure case for local paths: we just use the "path"
	// package's cleaning logic to remove any redundant "./" and "../"
	// sequences and any duplicate slashes and accept whatever that
	// produces.

	// Although using backslashes (Windows-style) is non-idiomatic, we do
	// allow it and just normalize it away, so the rest of Terraform will
	// only see the forward-slash form.
	if strings.Contains(raw, `\`) {
		// Note: We use string replacement rather than filepath.ToSlash
		// here because the filepath package behavior varies by current
		// platform, but we want to interpret configured paths the same
		// across all platforms: these are virtual paths within a module
		// package, not physical filesystem paths.
		raw = strings.ReplaceAll(raw, `\`, "/")
	}

	// Note that we could've historically blocked using "//" in a path here
	// in order to avoid confusion with the subdir syntax in remote addresses,
	// but we historically just treated that as the same as a single slash
	// and so we continue to do that now for compatibility. Clean strips those
	// out and reduces them to just a single slash.
	clean := path.Clean(raw)

	// However, we do need to keep a single "./" on the front if it isn't
	// a "../" path, or else it would be ambigous with the registry address
	// syntax.
	if !strings.HasPrefix(clean, "../") {
		clean = "./" + clean
	}

	return ModuleSourceLocal(clean), nil
}

func isModuleSourceLocal(raw string) bool {
	for _, prefix := range moduleSourceLocalPrefixes {
		if strings.HasPrefix(raw, prefix) {
			return true
		}
	}
	return false
}

func (s ModuleSourceLocal) moduleSource() {}

func (s ModuleSourceLocal) String() string {
	// We assume that our underlying string was already normalized at
	// construction, so we just return it verbatim.
	return string(s)
}

// ModuleSourceRemote is a ModuleSource representing a remote location from
// which we can retrieve a module package.
//
// Note that unlike Terraform, this also includes the address of the
// ModuleSourceRegistry equivalent. TFLint does not need to distinguish
// between ModuleSourceRemote and ModuleSourceRegistry,
// so they are all treated as ModuleSourceRemote.
type ModuleSourceRemote string

func (s ModuleSourceRemote) moduleSource() {}

func (s ModuleSourceRemote) String() string {
	// The remote source is not normalized and returns the input value as-is.
	return string(s)
}
