package pages

import (
	"errors"
	"net/http"
)

func ErrorPattern(authorizer PageAuthorizer, decoratorStrategy PageDecoratorStrategy) ErrorPatternFunc {
	if authorizer == nil {
		panic("error pattern: authorizer is required")
	}
	if decoratorStrategy == nil {
		panic("error pattern: decorator strategy is required")
	}

	return func(r *http.Request, status int, err error) string {
		e := err
	Loop:
		for {
			switch t := e.(type) {
			case interface{ Pattern() string }:
				return t.Pattern()
			case interface {
				Pattern(*http.Request, int, error) string
			}:
				return t.Pattern(r, status, err)
			case interface{ Unwrap() error }:
				e = t.Unwrap()
				continue
			default:
				break Loop
			}
		}

		if errors.Is(err, ErrPageNotFound) &&
			decoratorStrategy.IsURIDecorable(r.Context(), r.URL.Path) &&
			authorizer.Authorize(r.Context(), CreatePage) == Allow {
			return PageInternalCreate
		}

		switch status {
		case http.StatusUnauthorized:
			return PageErrorUnauthorized
		case http.StatusForbidden:
			return PageErrorForbidden
		case http.StatusNotFound:
			return PageErrorNotFound
		default:
			if status >= 400 && status < 500 {
				return PageError4xx
			}
			return PageError5xx
		}
	}
}
