package pages

import (
	"fmt"
	"maps"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/dlclark/regexp2"
	"golang.org/x/text/language"
)

var _ SiteRetriever = (*DefaultSiteRetriever)(nil)

const (
	headerAcceptLanguage = "Accept-Language"
	headerCFIPCountry    = "CF-IPCountry"
)

type SiteRetriever interface {
	Retrieve(r *http.Request) (*Site, string, error)
}

type DefaultSiteRetriever struct {
	store       SiteStore
	countryFunc func(*http.Request) (string, error)
	errorFunc   func(*http.Request, error) (*Site, error)
}

func NewDefaultSiteRetriever(
	store SiteStore,
	countryFunc func(*http.Request) (string, error),
	errorFunc func(*http.Request, error) (*Site, error),
) *DefaultSiteRetriever {
	if store == nil {
		panic("site retriever: store is required")
	}

	if countryFunc == nil {
		countryFunc = func(r *http.Request) (string, error) {
			return strings.ToUpper(r.Header.Get(headerCFIPCountry)), nil
		}
	}

	if errorFunc == nil {
		errorFunc = func(_ *http.Request, err error) (*Site, error) {
			return nil, err
		}
	}

	return &DefaultSiteRetriever{
		store:       store,
		countryFunc: countryFunc,
		errorFunc:   errorFunc,
	}
}

type candidate struct {
	site *Site
	path string
}

func (s *DefaultSiteRetriever) Retrieve(r *http.Request) (*Site, string, error) {
	if r == nil {
		panic("site retriever: request is required")
	}

	country, err := s.countryFunc(r)
	if err != nil {
		if site, pathInfo, err := s.resolveError(r, err); site != nil || err != nil {
			return site, pathInfo, err
		}
	}

	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	r.URL.RawPath = r.URL.Path

	host := getHost(r)

	localhosts := make([]candidate, 0, 1)
	defaults := make([]candidate, 0, 1)
	candidates := make([]candidate, 0, 5)
	sites := make([]candidate, 0, 10)

	for site, err := range s.store.FindPublished(r.Context()) {
		if err != nil {
			if site, pathInfo, err := s.resolveError(r, err); site != nil || err != nil {
				return site, pathInfo, err
			}
			continue
		}

		if site.IsLocalhost() {
			localhosts = append(localhosts, candidate{site: site})
		}

		if site.Host != host {
			continue
		}

		sites = append(sites, candidate{site: site})

		if site.IsDefault {
			defaults = append(defaults, candidate{site: site})
		}

		pathInfo, err := matchRequest(r, site.RelativePath)
		if err != nil {
			continue
		}

		candidates = append(candidates, candidate{site: site, path: pathInfo})
	}

	switch len(candidates) {
	case 0:
		if site, pathInfo := s.candidate(r, sites, country); site != nil {
			return site, pathInfo, nil
		}
	case 1:
		return candidates[0].site, candidates[0].path, nil
	}

	if site, pathInfo := s.candidate(r, candidates, country); site != nil {
		return site, pathInfo, nil
	}

	switch len(defaults) {
	case 0:
		if len(localhosts) > 0 {
			return localhosts[0].site, "", nil
		}
		return nil, "", ErrSiteNotFound
	case 1:
		return defaults[0].site, "", nil
	}

	if site, pathInfo := s.candidate(r, defaults, country); site != nil {
		return site, pathInfo, nil
	}
	return defaults[0].site, "", nil
}

func (s *DefaultSiteRetriever) candidate(r *http.Request, candidates []candidate, country string) (*Site, string) {
	var defaultCandidate candidate
	candidateTags := make(map[language.Tag]candidate)
	for _, c := range candidates {
		if country != "" && len(c.site.Countries) > 0 && !slices.Contains(c.site.Countries, country) {
			continue
		}
		if country != "" {
			if cTag, ok := candidateTags[c.site.Tag()]; ok && len(cTag.site.Countries) > 0 {
				continue
			}
		}
		defaultCandidate = c
		candidateTags[c.site.Tag()] = c
	}

	switch len(candidateTags) {
	case 1:
		return defaultCandidate.site, defaultCandidate.path
	case 0:
		return nil, ""
	}

	t, _, err := language.ParseAcceptLanguage(r.Header.Get(headerAcceptLanguage))
	if err != nil || len(t) == 0 {
		for _, c := range candidateTags {
			return c.site, c.path
		}
	}

	matcher := language.NewMatcher(slices.Collect(maps.Keys(candidateTags)))
	tag, _, _ := matcher.Match(t...)

	for !tag.IsRoot() {
		if c, ok := candidateTags[tag]; ok {
			return c.site, c.path
		}
		tag = tag.Parent()
	}

	return defaultCandidate.site, defaultCandidate.path
}

func (s *DefaultSiteRetriever) resolveError(r *http.Request, err error) (*Site, string, error) {
	site, err := s.errorFunc(r, err)
	if err != nil {
		return nil, "", err
	}

	if site == nil {
		return nil, "", nil
	}

	pathInfo, _ := matchRequest(r, site.RelativePath)

	return site, pathInfo, nil
}

const (
	reOptions    = regexp2.IgnoreCase & regexp2.RE2
	rePathExpr   = "^(%s)(/.*|$)"
	reNoPathExpr = "^()(/.*|$)"
)

var (
	reNoPath = regexp2.MustCompile(reNoPathExpr, reOptions)
	rePaths  = new(sync.Map)
)

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

func matchRequest(r *http.Request, relativePath string) (string, error) {
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

func getHost(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.Host); err == nil {
		return host
	}
	return r.Host
}
