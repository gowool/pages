package model

import (
	"maps"
	"net/http"
	"slices"

	"github.com/gowool/pages/internal"
)

var MultisiteStrategies = []MultisiteStrategy{Host, HostByLocale, HostWithPath, HostWithPathByLocale}

const (
	Host                 = MultisiteStrategy("host")
	HostByLocale         = MultisiteStrategy("host-by-locale")
	HostWithPath         = MultisiteStrategy("host-with-path")
	HostWithPathByLocale = MultisiteStrategy("host-with-path-by-locale")
)

type MultisiteStrategy string

func (t MultisiteStrategy) IsZero() bool {
	return t == ""
}

func (t MultisiteStrategy) String() string {
	return string(t)
}

type Configuration struct {
	Debug                 bool              `json:"debug,omitempty" yaml:"debug,omitempty" required:"true"`
	Multisite             MultisiteStrategy `json:"multisite,omitempty" yaml:"multisite,omitempty" required:"false" enum:"host,host-by-locale,host-with-path,host-with-path-by-locale"`
	FallbackLocale        string            `json:"fallbackLocale,omitempty" yaml:"fallbackLocale,omitempty" required:"false"`
	IgnoreRequestPatterns []string          `json:"ignoreRequestPatterns,omitempty" yaml:"ignoreRequestPatterns,omitempty" required:"false"`
	IgnoreRequestURIs     []string          `json:"ignoreRequestURIs,omitempty" yaml:"ignoreRequestURIs,omitempty" required:"false"`
	CatchErrors           map[string][]int  `json:"catchErrors,omitempty" yaml:"catchErrors,omitempty" required:"false"`
	Additional            map[string]string `json:"additional,omitempty" yaml:"additional,omitempty" required:"false"`
}

func NewConfiguration() Configuration {
	return Configuration{
		Multisite:      Host,
		FallbackLocale: "en_US",
		CatchErrors: map[string][]int{
			PageError4xx: {
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusPaymentRequired,
				http.StatusForbidden,
				http.StatusNotFound,
				http.StatusMethodNotAllowed,
				http.StatusNotAcceptable,
				http.StatusProxyAuthRequired,
				http.StatusRequestTimeout,
				http.StatusConflict,
				http.StatusGone,
				http.StatusLengthRequired,
				http.StatusPreconditionFailed,
				http.StatusRequestEntityTooLarge,
				http.StatusRequestURITooLong,
				http.StatusUnsupportedMediaType,
				http.StatusRequestedRangeNotSatisfiable,
				http.StatusExpectationFailed,
				http.StatusTeapot,
				http.StatusMisdirectedRequest,
				http.StatusUnprocessableEntity,
				http.StatusLocked,
				http.StatusFailedDependency,
				http.StatusTooEarly,
				http.StatusUpgradeRequired,
				http.StatusPreconditionRequired,
				http.StatusTooManyRequests,
				http.StatusRequestHeaderFieldsTooLarge,
				http.StatusUnavailableForLegalReasons,
			},
			PageError5xx: {
				http.StatusInternalServerError,
				http.StatusNotImplemented,
				http.StatusBadGateway,
				http.StatusServiceUnavailable,
				http.StatusGatewayTimeout,
				http.StatusHTTPVersionNotSupported,
				http.StatusVariantAlsoNegotiates,
				http.StatusInsufficientStorage,
				http.StatusLoopDetected,
				http.StatusNotExtended,
				http.StatusNetworkAuthenticationRequired,
			},
		},
		Additional: map[string]string{},
	}
}

func (c Configuration) IgnorePattern(pattern string) bool {
	if pattern == "" {
		return false
	}

	for _, expr := range c.IgnoreRequestPatterns {
		if re, ok := internal.Regexp(expr); ok {
			if ok, _ = re.MatchString(pattern); ok {
				return true
			}
		}
	}
	return false
}

func (c Configuration) IgnoreURI(uri string) bool {
	for _, expr := range c.IgnoreRequestURIs {
		if re, ok := internal.Regexp(expr); ok {
			if ok, _ = re.MatchString(uri); ok {
				return true
			}
		}
	}
	return false
}

func (c Configuration) With(other Configuration) Configuration {
	c.Debug = other.Debug

	if !other.Multisite.IsZero() {
		c.Multisite = other.Multisite
	}

	c.IgnoreRequestPatterns = internal.Unique(slices.Concat(c.IgnoreRequestPatterns, other.IgnoreRequestPatterns))
	c.IgnoreRequestURIs = internal.Unique(slices.Concat(c.IgnoreRequestURIs, other.IgnoreRequestURIs))

	if other.FallbackLocale != "" {
		c.FallbackLocale = other.FallbackLocale
	}

	if other.CatchErrors != nil {
		if c.CatchErrors == nil {
			c.CatchErrors = make(map[string][]int)
		}
		for pattern, codes := range other.CatchErrors {
			c.CatchErrors[pattern] = internal.Unique(slices.Concat(c.CatchErrors[pattern], codes))
		}
	}

	if other.Additional != nil {
		if c.Additional == nil {
			c.Additional = make(map[string]string)
		}
		maps.Copy(c.Additional, other.Additional)
	}
	return c
}
