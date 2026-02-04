package pages

import (
	"context"
	"net/http"
	"strings"

	"github.com/gowool/pages/internal"
)

const (
	HeaderXPageDecorable    = "X-Page-Decorable"
	HeaderXPageNotDecorable = "X-Page-Not-Decorable"
	headerXRequestedWith    = "X-Requested-With"
	xmlHTTPRequest          = "XMLHttpRequest"
)

type MiddlewareFunc func(Handler) Handler

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

type ErrorHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, error)
}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error)

func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, e error) {
	f(w, r, e)
}

type PageContext struct {
	*Context
	Request *http.Request
}

func (c PageContext) Value(key any) any {
	return c.Request.Context().Value(key)
}

func PageCtx() PageCtxFunc {
	return func(r *http.Request, c *Context) any {
		return PageContext{Context: c, Request: r}
	}
}

func ErrorPattern() ErrorPatternFunc {
	return func(_ context.Context, status int) string {
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

func PatternArgs() PatternArgsFunc {
	return func(r *http.Request) (args []any) {
		pattern := internal.Pattern(r)

		n := strings.Count(pattern, "{")
		if n == 0 {
			return
		}

		args = make([]any, 0, n*2)

		var key strings.Builder
		for _, c := range pattern {
			switch c {
			case '{':
				key.Reset()
				key.WriteRune(c)
			case '.':
				continue
			case '}':
				if key.Len() == 0 {
					panic("invalid dynamic page pattern")
				}
				key.WriteRune(c)
				param := key.String()
				args = append(args, param, r.PathValue(param[1:len(param)-1]))
			default:
				if key.Len() > 0 {
					key.WriteRune(c)
				}
			}
		}
		return
	}
}

func IsDecorable(w http.ResponseWriter, r *http.Request) bool {
	contentType := w.Header().Get(headerContentType)

	if contentType != "" && !strings.HasPrefix(contentType, mimeTextHTML) {
		return false
	}

	if w.Header().Get(HeaderXPageNotDecorable) == "1" {
		return false
	}

	if w.Header().Get(HeaderXPageDecorable) == "1" {
		return true
	}

	if r.Header.Get(headerXRequestedWith) == xmlHTTPRequest {
		return false
	}

	return ResponseStatus(w) == http.StatusOK
}

func ResponseStatus(w http.ResponseWriter) int {
	for {
		switch t := w.(type) {
		case interface{ Status() int }:
			return t.Status()
		case interface{ StatusCode() int }:
			return t.StatusCode()
		case interface{ Unwrap() http.ResponseWriter }:
			w = t.Unwrap()
			continue
		default:
			return 0
		}
	}
}
