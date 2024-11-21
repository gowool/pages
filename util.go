package pages

import (
	"mime"
	"net"
	"net/http"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/gowool/pages/internal"
)

const (
	headerAcceptLanguage = "Accept-Language"
	headerPageDecorate   = "X-Page-Decorate"
	xmlHTTPRequest       = "XMLHttpRequest"
)

func IsTLS(r *http.Request) bool {
	return r.TLS != nil
}

func Scheme(r *http.Request) string {
	if IsTLS(r) {
		return "https"
	}
	if scheme := r.Header.Get(echo.HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := r.Header.Get(echo.HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := r.Header.Get(echo.HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := r.Header.Get(echo.HeaderXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

func Host(r *http.Request) string {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		return r.Host
	}
	if port == "80" || port == "443" {
		return host
	}
	return host + ":" + port
}

func WithPageDecorate(w http.ResponseWriter, value string) {
	w.Header().Set(headerPageDecorate, value)
}

func PageDecorate(w http.ResponseWriter) bool {
	return w.Header().Get(headerPageDecorate) == "1"
}

func PageNotDecorate(w http.ResponseWriter) bool {
	return w.Header().Get(headerPageDecorate) == "0"
}

func IsAjax(r *http.Request) bool {
	return r.Header.Get(echo.HeaderXRequestedWith) == xmlHTTPRequest
}

func MediaType(header http.Header) string {
	ct, _, _ := mime.ParseMediaType(header.Get(echo.HeaderContentType))
	return ct
}

func IsTextHTML(header http.Header) bool {
	return MediaType(header) == echo.MIMETextHTML
}

type acceptLang struct {
	l string
	q string
}

func Languages(r *http.Request) []string {
	header := r.Header.Get(headerAcceptLanguage)

	languages := internal.Map(strings.Split(header, ","), func(item string) acceptLang {
		data := strings.Split(strings.TrimSpace(item), ";q=")
		codes := strings.Split(data[0], "-")

		if codes[0][0] == 'i' {
			// Language not listed in ISO 639 that are not variants
			// of any listed language, which can be registered with the
			// i-prefix, such as i-cherokee
			if len(codes[0]) > 1 {
				codes[0] = codes[0][1:]
			}
		}

		if len(codes) > 1 {
			codes[1] = strings.ToUpper(codes[1])
		}

		l := acceptLang{l: strings.Join(codes, "_"), q: "1.0"}

		if len(data) > 1 {
			l.q = data[1]
			return l
		}

		if data[0] == "*" {
			l.q = "0.0"
			return l
		}

		return l
	})

	sort.Slice(languages, func(i, j int) bool {
		return languages[i].q > languages[j].q
	})

	return internal.Map(languages, func(item acceptLang) string {
		return item.l
	})
}

func getLocale(r *http.Request, fallback string) string {
	if languages := Languages(r); len(languages) > 0 {
		return languages[0]
	}
	return fallback
}
