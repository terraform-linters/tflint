package formatter

import (
	"errors"

	hcl "github.com/hashicorp/hcl/v2"
)

type errorMapper[T any] struct {
	diagnostics func(error, hcl.Diagnostics) []T
	error       func(error) T
}

func mapErrors[T any](err error, mapper errorMapper[T]) []T {
	if err == nil {
		return []T{}
	}

	if errs, ok := err.(interface{ Unwrap() []error }); ok {
		var results []T
		for _, e := range errs.Unwrap() {
			results = append(results, mapErrors(e, mapper)...)
		}
		return results
	}

	var diags hcl.Diagnostics
	if errors.As(err, &diags) {
		return mapper.diagnostics(err, diags)
	}

	return []T{mapper.error(err)}
}
