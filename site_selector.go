package pages

import (
	"errors"
	"net/http"
)

var _ SiteSelector = (*DefaultSiteSelector)(nil)

const (
	headerXForwardedProto    = "X-Forwarded-Proto"
	headerXForwardedProtocol = "X-Forwarded-Protocol"
	headerXForwardedSsl      = "X-Forwarded-Ssl"
	headerXUrlScheme         = "X-Url-Scheme"
)

type SiteSelector interface {
	Select(r *http.Request) error
}

type DefaultSiteSelector struct {
	retriever SiteRetriever
}

func NewDefaultSiteSelector(retriever SiteRetriever) *DefaultSiteSelector {
	if retriever == nil {
		panic("site selector: retriever is required")
	}

	return &DefaultSiteSelector{retriever: retriever}
}

func (s *DefaultSiteSelector) Select(r *http.Request) error {
	c := FromContext(r.Context())
	if c == nil {
		panic("site selector: context is required")
	}

	site, pathInfo, err := s.retriever.Retrieve(r)
	if err != nil {
		if errors.Is(err, ErrSiteNotFound) {
			return err
		}
		return errors.Join(err, ErrSiteNotFound)
	}

	if site == nil {
		return ErrSiteNotFound
	}

	if pathInfo != "" {
		r.URL.Path = pathInfo
	}

	site.Host = r.Host
	site.Scheme = getScheme(r)

	c.SetSite(site)

	return nil
}

func getScheme(r *http.Request) string {
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
