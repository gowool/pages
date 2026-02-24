package pages

import (
	"net/http"
	"strings"

	"github.com/gowool/keratin"
)

const (
	HeaderXPageDecorable    = "X-Page-Decorable"
	HeaderXPageNotDecorable = "X-Page-Not-Decorable"
)

type PatternArgsFunc func(*http.Request) []any

func PatternArgs() PatternArgsFunc {
	return func(r *http.Request) (args []any) {
		pattern := keratin.Pattern(r)

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
	contentType := w.Header().Get(keratin.HeaderContentType)

	if contentType != "" && !strings.HasPrefix(contentType, keratin.MIMETextHTML) {
		return false
	}

	if w.Header().Get(HeaderXPageNotDecorable) == "1" {
		return false
	}

	if w.Header().Get(HeaderXPageDecorable) == "1" {
		return true
	}

	if r.Header.Get(keratin.HeaderXRequestedWith) == keratin.XMLHTTPRequest {
		return false
	}

	return keratin.ResponseStatusCode(w) == http.StatusOK
}
