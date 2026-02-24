package pages

import (
	"maps"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/spf13/cast"
)

var (
	DynamicPatternChars = "{"
	PageCMSPattern      = "/{_cms...}"
	HomeHybridPattern   = "/{$}"
)

const (
	PageCMS               = "_cms"
	PageAliasPrefix       = "_alias_"
	PageInternalPrefix    = "_internal_"
	PageInternalCreate    = PageInternalPrefix + "create"
	PageErrorPrefix       = PageInternalPrefix + "error_"
	PageErrorUnauthorized = PageErrorPrefix + "401"
	PageErrorForbidden    = PageErrorPrefix + "403"
	PageErrorNotFound     = PageErrorPrefix + "404"
	PageError4xx          = PageErrorPrefix + "4xx"
	PageError5xx          = PageErrorPrefix + "5xx"
)

type Page struct {
	ID ID `json:"id,omitempty" yaml:"id,omitempty"`

	SiteID ID    `json:"siteID,omitempty" yaml:"siteID,omitempty"`
	Site   *Site `json:"site,omitempty" yaml:"site,omitempty"`

	ParentID *ID     `json:"parentID,omitempty" yaml:"parentID,omitempty"`
	Parent   *Page   `json:"parent,omitempty" yaml:"parent,omitempty"`
	Children []*Page `json:"children,omitempty" yaml:"children,omitempty"`

	Created time.Time `json:"created,omitzero" yaml:"created,omitempty"`
	Updated time.Time `json:"updated,omitzero" yaml:"updated,omitempty"`

	Status     Status     `json:"status,omitempty" yaml:"status,omitempty"`
	Visibility Visibility `json:"visibility,omitempty" yaml:"visibility,omitempty"`

	Meta   Meta                `json:"meta,omitempty" yaml:"meta,omitempty"`
	DOM    DOM                 `json:"dom,omitempty" yaml:"dom,omitempty"`
	Header map[string][]string `json:"header,omitempty" yaml:"header,omitempty"`

	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Title     string `json:"title,omitempty" yaml:"title,omitempty"`
	Pattern   string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Alias     string `json:"alias,omitempty" yaml:"alias,omitempty"`
	Slug      string `json:"slug,omitempty" yaml:"slug,omitempty"`
	URL       string `json:"url,omitempty" yaml:"url,omitempty"`
	CustomURL string `json:"customURL,omitempty" yaml:"customURL,omitempty"`
	Template  string `json:"template,omitempty" yaml:"template,omitempty"`
	Position  int    `json:"position,omitempty" yaml:"position,omitempty"`
	Decorate  bool   `json:"decorate,omitempty" yaml:"decorate,omitempty"`
}

func NewPage() *Page {
	t := time.Now().UTC()

	return &Page{
		Created:    t,
		Updated:    t,
		Meta:       NewMeta(nil),
		Header:     make(map[string][]string),
		Status:     Draft,
		Visibility: Public,
	}
}

func (p *Page) String() string {
	if p.Name == "" {
		return "n/a"
	}
	return p.Name
}

func (p *Page) AbsURL(args ...any) string {
	if p.IsInternal() {
		return ""
	}

	var isDynamic bool

	path := p.Pattern
	if p.IsCMS() {
		path = p.URL
	} else {
		isDynamic = p.IsDynamic()
	}

	if path == "" || path == "/" {
		if p.Site == nil {
			return "/"
		}
		return p.Site.Home()
	}

	q := make(url.Values)
	for i := 0; i < len(args); i += 2 {
		key := cast.ToString(args[i])
		if len(key) == 0 {
			continue
		}

		value := cast.ToString(args[i+1])

		if isDynamic && len(key) > 2 && key[0] == '{' && key[len(key)-1] == '}' {
			path = strings.ReplaceAll(path, key, value)
			continue
		}

		q.Add(key, value)
	}

	var link strings.Builder
	if p.Site != nil {
		link.WriteString(strings.TrimRight(p.Site.URL(), "/"))
	}
	if path[0] != '/' {
		link.WriteRune('/')
	}
	link.WriteString(path)
	if len(q) > 0 {
		link.WriteRune('?')
		link.WriteString(q.Encode())
	}

	return link.String()
}

func (p *Page) IsInternal() bool {
	return strings.HasPrefix(p.Pattern, PageInternalPrefix)
}

func (p *Page) IsError() bool {
	return strings.HasPrefix(p.Pattern, PageErrorPrefix)
}

func (p *Page) IsHybrid() bool {
	return !p.IsCMS() && !p.IsInternal()
}

func (p *Page) IsCMS() bool {
	return PageCMS == p.Pattern
}

func (p *Page) IsDynamic() bool {
	return p.IsHybrid() && strings.ContainsAny(p.Pattern, DynamicPatternChars)
}

func (p *Page) SetAlias(alias string) {
	if !strings.HasPrefix(alias, PageAliasPrefix) {
		alias = PageAliasPrefix + alias
	}
	p.Alias = alias
}

func (p *Page) FixURL() {
	if p.IsInternal() {
		p.URL = ""
		return
	}

	if p.IsCMS() {
		if p.CustomURL == "/" && p.Parent == nil {
			p.URL = "/"
			p.fixChildren()
			return
		}

		if p.Slug == "" && p.Name != "" {
			p.Slug = slug.Make(p.Name)
		}

		pageURL := strings.TrimLeft(p.CustomURL, "/")
		if pageURL == "" {
			pageURL = strings.TrimLeft(p.Slug, "/")
		}

		if p.Parent == nil {
			p.URL = "/" + pageURL
		} else {
			base := p.Parent.URL
			if !strings.HasSuffix(base, "/") {
				base += "/"
			}

			p.URL = base + pageURL
		}
	}

	p.fixChildren()
}

func (p *Page) fixChildren() {
	for _, child := range p.Children {
		child.Parent = p
		child.FixURL()
	}
}

func (p *Page) Copy() *Page {
	page := *p
	page.Meta = maps.Clone(p.Meta)
	page.DOM = p.DOM.Copy()

	if p.ParentID != nil {
		page.ParentID = new(*p.ParentID)
	}

	if p.Parent != nil {
		page.Parent = p.Parent.Copy()
	}

	if p.Site != nil {
		page.Site = p.Site.Copy()
	}

	if p.Header != nil {
		page.Header = make(map[string][]string, len(p.Header))
		for k, v := range p.Header {
			page.Header[k] = slices.Clone(v)
		}
	}

	if len(p.Children) > 0 {
		page.Children = make([]*Page, len(p.Children))
		for i, child := range p.Children {
			page.Children[i] = child.Copy()
		}
	}

	return &page
}
