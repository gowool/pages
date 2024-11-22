package model

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
)

const (
	PageCMS            = "_page_cms"
	PageAliasPrefix    = "_page_alias_"
	PageInternalPrefix = "_page_internal_"
	PageInternalCreate = PageInternalPrefix + "create"
	PageErrorPrefix    = PageInternalPrefix + "error_"
	PageErrorInternal  = PageErrorPrefix + "internal"
	PageError4xx       = PageErrorPrefix + "4xx"
	PageError5xx       = PageErrorPrefix + "5xx"
)

type Page struct {
	ID          int64             `json:"id,omitempty" yaml:"id,omitempty" required:"true"`
	SiteID      int64             `json:"siteID,omitempty" yaml:"siteID,omitempty" required:"true"`
	ParentID    *int64            `json:"parentID,omitempty" yaml:"parentID,omitempty" required:"false"`
	Name        string            `json:"name,omitempty" yaml:"name,omitempty" required:"true"`
	Title       string            `json:"title,omitempty" yaml:"title,omitempty" required:"false"`
	Pattern     string            `json:"pattern,omitempty" yaml:"pattern,omitempty" required:"true"`
	Alias       string            `json:"alias,omitempty" yaml:"alias,omitempty" required:"false"`
	Slug        string            `json:"slug,omitempty" yaml:"slug,omitempty" required:"false"`
	URL         string            `json:"url,omitempty" yaml:"url,omitempty" required:"false"`
	CustomURL   string            `json:"customURL,omitempty" yaml:"customURL,omitempty" required:"false"`
	Javascript  string            `json:"javascript,omitempty" yaml:"javascript,omitempty" required:"false"`
	Stylesheet  string            `json:"stylesheet,omitempty" yaml:"stylesheet,omitempty" required:"false"`
	Template    string            `json:"template,omitempty" yaml:"template,omitempty" required:"true"`
	Decorate    bool              `json:"decorate,omitempty" yaml:"decorate,omitempty" required:"false"`
	Position    int               `json:"position,omitempty" yaml:"position,omitempty" required:"false"`
	Status      int               `json:"status,omitempty" yaml:"status,omitempty" required:"false"`
	ContentType string            `json:"contentType,omitempty" yaml:"contentType,omitempty" required:"false"`
	Headers     map[string]string `json:"headers,omitempty" yaml:"headers,omitempty" required:"false"`
	Metas       []Meta            `json:"metas,omitempty" yaml:"metas,omitempty" required:"false"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty" required:"false"`
	Created     time.Time         `json:"created,omitempty" yaml:"created,omitempty" required:"true"`
	Updated     time.Time         `json:"updated,omitempty" yaml:"updated,omitempty" required:"true"`
	Published   *time.Time        `json:"published,omitempty" yaml:"published,omitempty" required:"false"`
	Expired     *time.Time        `json:"expired,omitempty" yaml:"expired,omitempty" required:"false"`
	Site        *Site             `json:"-" yaml:"-"`
	Parent      *Page             `json:"-" yaml:"-"`
	Children    []Page            `json:"-" yaml:"-"`
}

func (p Page) GetID() int64 {
	return p.ID
}

func (p Page) String() string {
	if p.Name == "" {
		return "n/a"
	}
	return p.Name
}

func (p Page) IsEnabled(now time.Time) bool {
	now = now.Truncate(60 * time.Second)
	return p.Published != nil &&
		!p.Published.IsZero() &&
		(p.Published.Before(now) || p.Published.Equal(now)) &&
		(p.Expired == nil || p.Expired.IsZero() || p.Expired.After(now))
}

func (p Page) WithAlias(alias string) Page {
	if !strings.HasPrefix(alias, PageAliasPrefix) {
		alias = PageAliasPrefix + alias
	}
	p.Alias = alias
	return p
}

func (p Page) WithInternal(pattern string) Page {
	if !strings.HasPrefix(pattern, PageInternalPrefix) {
		pattern = PageAliasPrefix + pattern
	}
	p.Pattern = pattern
	return p
}

func (p Page) WithError(pattern string) Page {
	if !strings.HasPrefix(pattern, PageErrorPrefix) {
		pattern = PageErrorPrefix + pattern
	}
	p.Pattern = pattern
	return p
}

func (p Page) IsInternal() bool {
	return strings.HasPrefix(p.Pattern, PageInternalPrefix)
}

func (p Page) IsError() bool {
	return strings.HasPrefix(p.Pattern, PageErrorPrefix)
}

func (p Page) IsHybrid() bool {
	return !p.IsCMS() && !p.IsInternal()
}

func (p Page) IsCMS() bool {
	return PageCMS == p.Pattern
}

func (p Page) IsDynamic() bool {
	return p.IsHybrid() && strings.ContainsAny(p.URL, ":{*")
}

func (p Page) WithFixedURL() Page {
	if p.IsInternal() {
		p.URL = ""
		return p
	}

	if !p.IsHybrid() {
		if p.Parent == nil {
			p.Slug = ""

			p.URL = "/" + strings.TrimLeft(p.CustomURL, "/")
		} else {
			if p.Slug == "" {
				p.Slug = slug.Make(p.Name)
			}

			base := p.Parent.URL
			if !strings.HasSuffix(base, "/") {
				base += "/"
			}

			url := p.CustomURL
			if url == "" {
				url = p.Slug
			}

			p.URL = base + strings.TrimLeft(url, "/")
		}
	}

	children := make([]Page, 0, len(p.Children))
	for _, child := range p.Children {
		child.Parent = &p
		children = append(children, child.WithFixedURL())
	}
	p.Children = children
	return p
}
