package pages

import (
	"fmt"
	"net"
	"net/http"
	"slices"
	"sync"

	"github.com/dlclark/regexp2"
	"github.com/gowool/wo"
	"golang.org/x/text/language"
)

var _ SiteSelector = (*DefaultSiteSelector)(nil)

type SiteSelector interface {
	Retrieve(r *http.Request) (*Site, string, error)
}

type DefaultSiteSelector struct {
	store       SiteStore
	countryFunc func(*http.Request) (string, error)
	errorFunc   func(*http.Request, error) (*Site, error)
}

func NewSiteSelector(
	store SiteStore,
	countryFunc func(*http.Request) (string, error),
	errorFunc func(*http.Request, error) (*Site, error),
) *DefaultSiteSelector {
	if store == nil {
		panic("site selector: store is required")
	}

	if countryFunc == nil {
		countryFunc = func(r *http.Request) (string, error) {
			return r.Header.Get(wo.HeaderCFIPCountry), nil
		}
	}

	if errorFunc == nil {
		errorFunc = func(_ *http.Request, err error) (*Site, error) {
			return nil, err
		}
	}

	return &DefaultSiteSelector{
		store:       store,
		countryFunc: countryFunc,
		errorFunc:   errorFunc,
	}
}

func (s *DefaultSiteSelector) Retrieve(r *http.Request) (*Site, string, error) {
	if r == nil {
		panic("site selector: request is required")
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

	var (
		defaultSite *Site
		countrySite *Site
		matcherTags []language.Tag
	)

	host := getHost(r)
	sites := make(map[language.Tag]*Site)
	paths := make(map[language.Tag]string)

	for site, err := range s.store.FindEnabled(r.Context()) {
		if err != nil {
			if site, pathInfo, err := s.resolveError(r, err); site != nil || err != nil {
				return site, pathInfo, err
			}
			continue
		}

		if site.Host != host && !site.IsLocalhost() {
			continue
		}

		if site.IsDefault && (defaultSite == nil || defaultSite.IsLocalhost()) {
			defaultSite = site
		}

		var pi string

		if r.URL.RawPath == "/" {
			if country != "" && len(site.Countries) > 0 && !slices.Contains(site.Countries, country) {
				continue
			}
		} else {
			if (countrySite == nil || countrySite.IsLocalhost()) && country != "" && slices.Contains(site.Countries, country) {
				countrySite = site
			}

			var err1 error
			if pi, err1 = matchRequest(r, site.RelativePath); err1 != nil {
				continue
			}
		}

		tag := site.Tag()
		matcherTags = append(matcherTags, tag)

		if currentSite, ok := sites[tag]; !ok || (country != "" && !slices.Contains(currentSite.Countries, country)) {
			sites[tag] = site
			paths[tag] = pi
		}
	}

	selectedSite, pathInfo := s.selectedSite(r, sites, paths, matcherTags)

	if selectedSite == nil || selectedSite.IsLocalhost() {
		if countrySite == nil || countrySite.IsLocalhost() {
			selectedSite = defaultSite
		} else {
			selectedSite = countrySite
		}
	}

	if selectedSite != nil {
		return selectedSite, pathInfo, nil
	}

	if site, pathInfo, err := s.resolveError(r, ErrSiteNotFound); site != nil || err != nil {
		return site, pathInfo, err
	}

	return nil, "", ErrSiteNotFound
}

func (s *DefaultSiteSelector) resolveError(r *http.Request, err error) (*Site, string, error) {
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

func (s *DefaultSiteSelector) selectedSite(
	r *http.Request,
	sites map[language.Tag]*Site,
	paths map[language.Tag]string,
	matcherTags []language.Tag,
) (*Site, string) {
	if len(matcherTags) == 0 {
		return nil, ""
	}

	t, _, err := language.ParseAcceptLanguage(r.Header.Get(wo.HeaderAcceptLanguage))
	if err != nil || len(t) == 0 {
		t = []language.Tag{matcherTags[0]}
	}

	matcher := language.NewMatcher(matcherTags)
	tag, _, _ := matcher.Match(t...)
	for !tag.IsRoot() {
		if site, ok := sites[tag]; ok {
			return site, paths[tag]
		}
		tag = tag.Parent()
	}

	return nil, ""
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
