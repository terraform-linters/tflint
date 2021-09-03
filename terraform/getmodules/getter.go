package getmodules

import (
	getter "github.com/hashicorp/go-getter"
)

// We configure our own go-getter detector and getter sets here, because
// the set of sources we support is part of Terraform's documentation and
// so we don't want any new sources introduced in go-getter to sneak in here
// and work even though they aren't documented. This also insulates us from
// any meddling that might be done by other go-getter callers linked into our
// executable.
//
// Note that over time we've found go-getter's design to be not wholly fit
// for Terraform's purposes in various ways, and so we're continuing to use
// it here because our backward compatibility with earlier versions depends
// on it, but we use go-getter very carefully and always only indirectly via
// the public API of this package so that we can get the subset of the
// go-getter functionality we need while working around some of the less
// helpful parts of its design. See the comments in various other functions
// in this package which call into go-getter for more information on what
// tradeoffs we're making here.

var goGetterDetectors = []getter.Detector{
	new(getter.GitHubDetector),
	new(getter.GitDetector),

	// Because historically BitBucket supported both Git and Mercurial
	// repositories but used the same repository URL syntax for both,
	// this detector takes the unusual step of actually reaching out
	// to the BitBucket API to recognize the repository type. That
	// means there's the possibility of an outgoing network request
	// inside what is otherwise normally just a local string manipulation
	// operation, but we continue to accept this for now.
	//
	// Perhaps a future version of go-getter will remove the check now
	// that BitBucket only supports Git anyway. Aside from this historical
	// exception, we should avoid adding any new detectors that make network
	// requests in here, and limit ourselves only to ones that can operate
	// entirely through local string manipulation.
	new(getter.BitBucketDetector),

	new(getter.GCSDetector),
	new(getter.S3Detector),
	new(fileDetector),
}
