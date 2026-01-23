package pages

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/gowool/wo"
)

var _ Resolver = (*Event)(nil)

var ErrThemeRequired = errors.New("theme required")

const (
	HeaderXPageDecorable    = "X-Page-Decorable"
	HeaderXPageNotDecorable = "X-Page-Not-Decorable"
)

type Resolver interface {
	wo.Resolver

	// Scheme returns the HTTP protocol scheme, `http` or `https`.
	Scheme() string
	BindBody(dst any) error

	Site() *Site
	SetSite(site *Site)
	HasSite() bool

	Page() *Page
	SetPage(page *Page)
	HasPage() bool

	Status() int
	SetStatus(status int)

	SetError(err error)
	SetContent(content template.HTML)

	IsGuest() bool
	IsDecorable() bool

	NoContent(status int) error
	Render(status int, contentType, template string) error
}

type (
	keyStatus  struct{}
	keyError   struct{}
	keyContent struct{}
	keySite    struct{}
	keyPage    struct{}
	keyGuest   struct{}
)

type Event struct {
	wo.Event

	seo   *SEO
	theme Theme
}

func (e *Event) Reset(w http.ResponseWriter, r *http.Request, t Theme) {
	e.Event.Reset(w, r)
	e.SEO().Reset()
	e.theme = t
}

func (e *Event) SetTheme(t Theme) {
	e.theme = t
}

func (e *Event) Theme() Theme {
	return e.theme
}

func (e *Event) IsRoot() bool {
	return e.Request().URL.RawPath == "/"
}

func (e *Event) SEO() *SEO {
	if e.seo == nil {
		e.seo = NewSEO()
	}
	return e.seo
}

func (e *Event) Pattern() string {
	return getPattern(e.Request())
}

func (e *Event) IsGuest() bool {
	auth, _ := e.Value(keyGuest{}).(bool)
	return !auth
}

func (e *Event) SetGuest(guest bool) {
	e.SetValue(keyGuest{}, guest)
}

func (e *Event) Site() *Site {
	site, _ := e.Value(keySite{}).(*Site)
	return site
}

func (e *Event) SetSite(site *Site) {
	if site != nil {
		site.isRoot = e.IsRoot()
	}
	e.SEO().Site(site)
	e.SetValue(keySite{}, site)
}

func (e *Event) HasSite() bool {
	return e.Site() != nil
}

func (e *Event) Page() *Page {
	page, _ := e.Value(keyPage{}).(*Page)
	return page
}

func (e *Event) SetPage(page *Page) {
	var args []any

	if page != nil {
		if page.Site == nil {
			page.Site = e.Site()
		}

		if page.IsDynamic() {
			args = patternArgs(e.Param, e.Pattern())
		}
	}

	e.SEO().Page(page, args...)
	e.SetValue(keyPage{}, page)
}

func (e *Event) HasPage() bool {
	return e.Page() != nil
}

func (e *Event) Status() int {
	if status, ok := e.Value(keyStatus{}).(int); ok {
		return status
	}
	return http.StatusOK
}

func (e *Event) SetStatus(status int) {
	e.SetValue(keyStatus{}, status)
}

func (e *Event) Error() error {
	err, _ := e.Value(keyError{}).(error)
	return err
}

func (e *Event) SetError(err error) {
	e.SetValue(keyError{}, err)
}

func (e *Event) Content() template.HTML {
	if content, ok := e.Value(keyContent{}).(template.HTML); ok {
		return content
	}
	return ""
}

func (e *Event) SetContent(content template.HTML) {
	e.SetValue(keyContent{}, content)
}

func (e *Event) IsDecorable() bool {
	contentType := e.Response().Header().Get(wo.HeaderContentType)

	if contentType != "" && !strings.HasPrefix(contentType, wo.MIMETextHTML) {
		return false
	}

	if e.Response().Header().Get(HeaderXPageNotDecorable) == "1" {
		return false
	}

	if e.Response().Header().Get(HeaderXPageDecorable) == "1" {
		return true
	}

	if wo.MustUnwrapResponse(e.Response()).Status != http.StatusOK {
		return false
	}

	return !e.IsAjax()
}

func getPattern(r *http.Request) string {
	pattern := r.Pattern
	if index := strings.IndexRune(pattern, ' '); index > -1 {
		pattern = pattern[index+1:]
	}
	return pattern
}

func patternArgs(paramFunc func(string) string, pattern string) (args []any) {
	n := strings.Count(pattern, "{")
	if n == 0 {
		return
	}

	args = make([]any, 0, n*2)

	var key strings.Builder
	for _, c := range pattern {
		switch c {
		case '{':
			key = strings.Builder{}
			key.WriteRune(c)
		case '.':
			continue
		case '}':
			if key.Len() == 0 {
				panic("invalid dynamic page pattern")
			}
			key.WriteRune(c)
			param := key.String()
			args = append(args, param, paramFunc(param[1:len(param)-1]))
		default:
			if key.Len() > 0 {
				key.WriteRune(c)
			}
		}
	}
	return
}

func (e *Event) View(template string, data any) ([]byte, error) {
	if e.theme == nil {
		return nil, ErrThemeRequired
	}

	var buf bytes.Buffer
	if err := e.theme.Write(e.Context(), &buf, template, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *Event) Render(status int, contentType, template string) error {
	data, err := e.View(template, e)
	if err != nil {
		return err
	}
	return e.Blob(status, contentType, data)
}

func (e *Event) RenderHTML(status int, template string) error {
	return e.Render(status, wo.MIMETextHTMLCharsetUTF8, template)
}
