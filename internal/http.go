package internal

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/dlclark/regexp2"
)

const (
	headerXForwardedProto    = "X-Forwarded-Proto"
	headerXForwardedProtocol = "X-Forwarded-Protocol"
	headerXForwardedSsl      = "X-Forwarded-Ssl"
	headerXUrlScheme         = "X-Url-Scheme"

	reOptions    = regexp2.IgnoreCase & regexp2.RE2
	rePathExpr   = "^(%s)(/.*|$)"
	reNoPathExpr = "^()(/.*|$)"
)

var (
	reMethod = regexp.MustCompile(`^(\S*)\s+(.*)$`)
	reNoPath = regexp2.MustCompile(reNoPathExpr, reOptions)
	rePaths  = new(sync.Map)
)

func Host(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.Host); err == nil {
		return host
	}
	return r.Host
}

func Scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if scheme := r.Header.Get(headerXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := r.Header.Get(headerXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := r.Header.Get(headerXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := r.Header.Get(headerXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

func Pattern(r *http.Request) string {
	pattern := r.Pattern
	if index := strings.IndexRune(pattern, ' '); index > -1 {
		pattern = pattern[index+1:]
	}
	return pattern
}

func CheckMethod(method, skip string) (string, bool) {
	if matches := reMethod.FindStringSubmatch(skip); len(matches) > 2 {
		if matches[1] == method {
			return matches[2], true
		}
		return "", false
	}
	return skip, true
}

func MatchRequest(r *http.Request, relativePath string) (string, error) {
	var (
		re    *regexp2.Regexp
		match *regexp2.Match
		err   error
	)

	if relativePath == "" || relativePath == "/" {
		re = reNoPath
	} else if re, err = regexpPath(relativePath); err != nil {
		return "", err
	}

	if match, err = re.FindStringMatch(r.URL.Path); err != nil {
		return "", err
	}

	if match == nil {
		return "", fmt.Errorf("invalid path %s", r.URL.Path)
	}

	groups := match.Groups()

	if len(groups) < 3 {
		return "", fmt.Errorf("invalid match path %s", r.URL.Path)
	}

	matched := groups[2].String()
	if matched == "" {
		return "/", nil
	}
	return matched, nil
}

func regexpPath(path string) (*regexp2.Regexp, error) {
	if v, ok := rePaths.Load(path); ok {
		switch v := v.(type) {
		case *regexp2.Regexp:
			return v, nil
		case error:
			return nil, v
		}
	}

	re, err := regexp2.Compile(fmt.Sprintf(rePathExpr, path), reOptions)
	if err != nil {
		rePaths.Store(path, err)
		return nil, err
	}

	rePaths.Store(path, re)
	return re, nil
}
