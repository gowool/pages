package model

import (
	"maps"
	"net/http"
	"slices"
	"sync"

	"github.com/gowool/echox"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

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

type Skippers struct {
	EqualPaths  []string `json:"equalPaths,omitempty" yaml:"equalPaths,omitempty"`
	PrefixPaths []string `json:"prefixPaths,omitempty" yaml:"prefixPaths,omitempty"`
	SuffixPaths []string `json:"suffixPaths,omitempty" yaml:"suffixPaths,omitempty"`
	Expressions []string `json:"expressions,omitempty" yaml:"expressions,omitempty"`

	once    sync.Once
	skipper middleware.Skipper
}

func (s *Skippers) Skipper(c echo.Context) bool {
	s.once.Do(func() {
		s.skipper = echox.ChainSkipper(
			echox.EqualPathSkipper(s.EqualPaths...),
			echox.PrefixPathSkipper(s.PrefixPaths...),
			echox.SuffixPathSkipper(s.SuffixPaths...),
			echox.ExpressionSkipper(s.Expressions...),
		)
	})
	return s.skipper(c)
}

type Configuration struct {
	Debug                 bool              `json:"debug,omitempty" yaml:"debug,omitempty" required:"true"`
	Multisite             MultisiteStrategy `json:"multisite,omitempty" yaml:"multisite,omitempty" required:"false" enum:"host,host-by-locale,host-with-path,host-with-path-by-locale"`
	FallbackLocale        string            `json:"fallbackLocale,omitempty" yaml:"fallbackLocale,omitempty" required:"false"`
	IgnoreRequestPatterns []string          `json:"ignoreRequestPatterns,omitempty" yaml:"ignoreRequestPatterns,omitempty" required:"false"`
	IgnoreRequestURIs     []string          `json:"ignoreRequestURIs,omitempty" yaml:"ignoreRequestURIs,omitempty" required:"false"`
	SiteSkippers          *Skippers         `json:"siteSkippers,omitempty" yaml:"siteSkippers,omitempty" required:"false"`
	PageSkippers          *Skippers         `json:"pageSkippers,omitempty" yaml:"pageSkippers,omitempty" required:"false"`
	LoggerSkippers        *Skippers         `json:"loggerSkippers,omitempty" yaml:"loggerSkippers,omitempty" required:"false"`
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

	if other.FallbackLocale != "" {
		c.FallbackLocale = other.FallbackLocale
	}

	c.IgnoreRequestPatterns = concat(c.IgnoreRequestPatterns, other.IgnoreRequestPatterns)
	c.IgnoreRequestURIs = concat(c.IgnoreRequestURIs, other.IgnoreRequestURIs)

	c.SiteSkippers = skippersMerge(c.SiteSkippers, other.SiteSkippers)
	c.PageSkippers = skippersMerge(c.PageSkippers, other.PageSkippers)
	c.LoggerSkippers = skippersMerge(c.LoggerSkippers, other.LoggerSkippers)

	if other.CatchErrors != nil {
		if c.CatchErrors == nil {
			c.CatchErrors = make(map[string][]int)
		}
		for pattern, codes := range other.CatchErrors {
			c.CatchErrors[pattern] = concat(c.CatchErrors[pattern], codes)
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

func skippersMerge(a, b *Skippers) *Skippers {
	return &Skippers{
		EqualPaths:  concat(a.EqualPaths, b.EqualPaths),
		PrefixPaths: concat(a.PrefixPaths, b.PrefixPaths),
		SuffixPaths: concat(a.SuffixPaths, b.SuffixPaths),
		Expressions: concat(a.Expressions, b.Expressions),
	}
}

func concat[T comparable](a, b []T) []T {
	return internal.Unique(slices.Concat(a, b))
}
