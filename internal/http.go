package internal

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/dlclark/regexp2"
)

const (
	reOptions    = regexp2.IgnoreCase & regexp2.RE2
	rePathExpr   = "^(%s)(/.*|$)"
	reNoPathExpr = "^()(/.*|$)"
)

var (
	reNoPath = regexp2.MustCompile(reNoPathExpr, reOptions)
	rePaths  = new(sync.Map)
)

func Host(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.Host); err == nil {
		return host
	}
	return r.Host
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
