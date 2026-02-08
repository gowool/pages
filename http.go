package pages

import (
	"net/http"
	"strings"

	"github.com/gowool/gor"
	"github.com/gowool/gor/middleware"
)

const (
	HeaderXPageDecorable    = "X-Page-Decorable"
	HeaderXPageNotDecorable = "X-Page-Not-Decorable"
)

type (
	Handler          = gor.Handler
	HandlerFunc      = gor.HandlerFunc
	ErrorHandler     = gor.ErrorHandler
	ErrorHandlerFunc = gor.ErrorHandlerFunc
	Skipper          = middleware.Skipper
)

func Scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if scheme := r.Header.Get(gor.HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := r.Header.Get(gor.HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := r.Header.Get(gor.HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := r.Header.Get(gor.HeaderXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

func CheckMethod(method, pattern string) (string, bool) {
	if index := strings.IndexRune(pattern, ' '); index > 0 {
		if method == pattern[:index] {
			return strings.TrimSpace(pattern[index+1:]), true
		}
		return "", false
	}
	return pattern, true
}

func Pattern(r *http.Request) string {
	pattern := r.Pattern
	if index := strings.IndexRune(pattern, ' '); index > -1 {
		pattern = pattern[index+1:]
	}
	return pattern
}

func PatternArgs() PatternArgsFunc {
	return func(r *http.Request) (args []any) {
		pattern := Pattern(r)

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
	contentType := w.Header().Get(gor.HeaderContentType)

	if contentType != "" && !strings.HasPrefix(contentType, gor.MIMETextHTML) {
		return false
	}

	if w.Header().Get(HeaderXPageNotDecorable) == "1" {
		return false
	}

	if w.Header().Get(HeaderXPageDecorable) == "1" {
		return true
	}

	if r.Header.Get(gor.HeaderXRequestedWith) == gor.XMLHTTPRequest {
		return false
	}

	return gor.ResponseStatusCode(w) == http.StatusOK
}
