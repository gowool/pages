package pages

import (
	"net/http"
	"strings"

	"github.com/gowool/pages/internal"
)

type Skipper func(*http.Request) bool

func ChainSkipper(skippers ...Skipper) Skipper {
	return func(r *http.Request) bool {
		for _, skipper := range skippers {
			if skipper(r) {
				return true
			}
		}
		return false
	}
}

func PrefixPathSkipper(prefixes ...string) Skipper {
	prefixes = internal.Map(prefixes, strings.ToLower)
	return func(r *http.Request) bool {
		p := strings.ToLower(r.URL.Path)
		m := strings.ToLower(r.Method)
		for _, prefix := range prefixes {
			if prefix, ok := internal.CheckMethod(m, prefix); ok && strings.HasPrefix(p, prefix) {
				return true
			}
		}
		return false
	}
}

func SuffixPathSkipper(suffixes ...string) Skipper {
	suffixes = internal.Map(suffixes, strings.ToLower)
	return func(r *http.Request) bool {
		p := strings.ToLower(r.URL.Path)
		m := strings.ToLower(r.Method)
		for _, suffix := range suffixes {
			if suffix, ok := internal.CheckMethod(m, suffix); ok && strings.HasSuffix(p, suffix) {
				return true
			}
		}
		return false
	}
}

func EqualPathSkipper(paths ...string) Skipper {
	return func(r *http.Request) bool {
		for _, path := range paths {
			if path, ok := internal.CheckMethod(r.Method, path); ok && strings.EqualFold(r.URL.Path, path) {
				return true
			}
		}
		return false
	}
}

func PageSkipper(decoratorStrategy PageDecoratorStrategy) Skipper {
	if decoratorStrategy == nil {
		panic("page skipper: decorator strategy is required")
	}

	return func(r *http.Request) bool {
		if r.Method != http.MethodGet {
			return true
		}

		if pattern := internal.Pattern(r); pattern != PageCMSPattern {
			return !decoratorStrategy.IsPatternDecorable(r.Context(), pattern)
		}

		return !decoratorStrategy.IsURIDecorable(r.Context(), r.URL.Path)
	}
}
