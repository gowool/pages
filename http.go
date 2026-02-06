package pages

import (
	"net/http"
	"strings"
)

const (
	HeaderAccept             = "Accept"
	HeaderAcceptLanguage     = "Accept-Language"
	HeaderContentType        = "Content-Type"
	HeaderCFIPCountry        = "CF-IPCountry"
	HeaderXPageDecorable     = "X-Page-Decorable"
	HeaderXPageNotDecorable  = "X-Page-Not-Decorable"
	HeaderXRequestedWith     = "X-Requested-With"
	HeaderXForwardedProto    = "X-Forwarded-Proto"
	HeaderXForwardedProtocol = "X-Forwarded-Protocol"
	HeaderXForwardedSsl      = "X-Forwarded-Ssl"
	HeaderXUrlScheme         = "X-Url-Scheme"
	XMLHTTPRequest           = "XMLHttpRequest"
	MIMEApplicationJSON      = "application/json"
	MIMETextHTML             = "text/html"
	MIMETextHTMLCharsetUTF8  = MIMETextHTML + "; charset=UTF-8"
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

func Scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if scheme := r.Header.Get(HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := r.Header.Get(HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := r.Header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := r.Header.Get(HeaderXUrlScheme); scheme != "" {
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
	contentType := w.Header().Get(HeaderContentType)

	if contentType != "" && !strings.HasPrefix(contentType, MIMETextHTML) {
		return false
	}

	if w.Header().Get(HeaderXPageNotDecorable) == "1" {
		return false
	}

	if w.Header().Get(HeaderXPageDecorable) == "1" {
		return true
	}

	if r.Header.Get(HeaderXRequestedWith) == XMLHTTPRequest {
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

func ResponseSize(w http.ResponseWriter) int64 {
	for {
		switch t := w.(type) {
		case interface{ Size() int64 }:
			return t.Size()
		case interface{ Unwrap() http.ResponseWriter }:
			w = t.Unwrap()
			continue
		default:
			return -1
		}
	}
}

func ResponseCommitted(w http.ResponseWriter) bool {
	for {
		switch t := w.(type) {
		case interface{ Committed() bool }:
			return t.Committed()
		case interface{ Unwrap() http.ResponseWriter }:
			w = t.Unwrap()
			continue
		default:
			return false
		}
	}
}
