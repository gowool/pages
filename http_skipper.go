package pages

import (
	"net/http"

	"github.com/gowool/keratin/middleware"
)

func PageSkipper(decoratorStrategy PageDecoratorStrategy) middleware.Skipper {
	if decoratorStrategy == nil {
		panic("page skipper: decorator strategy is required")
	}

	return func(r *http.Request) bool {
		if r.Method != http.MethodGet {
			return true
		}

		if pattern := Pattern(r); pattern != PageCMSPattern {
			return !decoratorStrategy.IsPatternDecorable(r.Context(), pattern)
		}

		return !decoratorStrategy.IsURIDecorable(r.Context(), r.URL.Path)
	}
}
