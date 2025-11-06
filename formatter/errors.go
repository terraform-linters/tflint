package formatter

import (
	"errors"

	hcl "github.com/hashicorp/hcl/v2"
)

type errorMapper[T any] struct {
	diagnostic func(*hcl.Diagnostic) T
	error      func(error) T
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
		results := make([]T, len(diags))
		for i, diag := range diags {
			results[i] = mapper.diagnostic(diag)
		}
		return results
	}

	return []T{mapper.error(err)}
}
