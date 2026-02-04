package pages

import (
	"context"
	"html/template"
	"net/http"
	"sync"
)

type contextKey struct{}

func FromContext(ctx context.Context) *Context {
	c, _ := ctx.Value(contextKey{}).(*Context)
	return c
}

var ctxPool = &sync.Pool{
	New: func() any {
		c := new(Context)
		c.Reset()
		return c
	},
}

func NewContext(parent context.Context) (context.Context, context.CancelFunc) {
	c := ctxPool.Get().(*Context)

	cancel := func() {
		c.Reset()
		ctxPool.Put(c)
	}

	return context.WithValue(parent, contextKey{}, c), cancel
}

type Context struct {
	seo     *SEO
	site    *Site
	page    *Page
	err     error
	debug   bool
	guest   bool
	status  int
	content template.HTML
}

func (c *Context) Reset() {
	c.SEO().Reset()
	c.status = http.StatusOK
	c.content = ""
	c.debug = false
	c.guest = true
	c.site = nil
	c.page = nil
	c.err = nil
}

func (c *Context) SEO() *SEO {
	if c.seo == nil {
		c.seo = NewSEO()
	}
	return c.seo
}

func (c *Context) Site() *Site {
	return c.site
}

func (c *Context) SetSite(site *Site) {
	if site != nil {
		c.SEO().Site(site)
	}
	c.site = site
}

func (c *Context) HasSite() bool {
	return c.site != nil
}

func (c *Context) Page() *Page {
	return c.page
}

func (c *Context) SetPage(page *Page, args ...any) {
	if page != nil {
		if page.Site == nil && c.HasSite() {
			page.SiteID = c.site.ID
			page.Site = c.site
		}
		c.SEO().Page(page, args...)
	}
	c.page = page
}

func (c *Context) HasPage() bool {
	return c.page != nil
}

func (c *Context) Error() error {
	return c.err
}

func (c *Context) SetError(err error) {
	c.err = err
}

func (c *Context) HasError() bool {
	return c.err != nil
}

func (c *Context) Debug() bool {
	return c.debug
}

func (c *Context) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Context) Guest() bool {
	return c.guest
}

func (c *Context) SetGuest(guest bool) {
	c.guest = guest
}

func (c *Context) Status() int {
	if c.status == 0 {
		return http.StatusOK
	}
	return c.status
}

func (c *Context) SetStatus(status int) {
	c.status = status
}

func (c *Context) Content() template.HTML {
	return c.content
}

func (c *Context) SetContent(content template.HTML) {
	c.content = content
}

func (c *Context) HasContent() bool {
	return c.content != ""
}
