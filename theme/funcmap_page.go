package theme

import (
	"context"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"time"

	"github.com/gowool/cr"
	"github.com/gowool/theme"

	"github.com/gowool/pages"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type FuncMapPage struct {
	pageRepo repository.Page
}

func NewFuncMapPage(pageRepo repository.Page) *FuncMapPage {
	return &FuncMapPage{
		pageRepo: pageRepo,
	}
}

func (fm *FuncMapPage) FuncMap(t theme.Theme) template.FuncMap {
	return template.FuncMap{
		"page_url":          fm.pageURL,
		"page_by_id":        fm.findPage,
		"page_children":     fm.pageChildren,
		"pages_by_criteria": fm.pagesByCriteria,
	}
}

func (fm *FuncMapPage) pageURL(ctx context.Context, name any, args ...any) string {
	switch name := name.(type) {
	case string:
		if strings.HasPrefix(name, model.PageAliasPrefix) {
			return fm.pageURLByAlias(ctx, name, args...)
		}
		if name == model.PageCMS {
			return fm.pageURLByPath(ctx, args...)
		}
	case model.Page:
		return fm.pageURLByPage(ctx, name, args...)
	case *model.Page:
		return fm.pageURLByPage(ctx, *name, args...)
	}
	return ""
}

func (fm *FuncMapPage) pageURLByAlias(ctx context.Context, alias string, args ...any) string {
	site := pages.CtxSite(ctx)
	if site == nil {
		return ""
	}
	page, err := fm.pageRepo.FindByAlias(ctx, site.ID, alias, time.Now())
	if err != nil {
		return ""
	}
	page.Site = site
	return fm.pageURLByPage(ctx, page, args...)
}

func (fm *FuncMapPage) pageURLByPath(ctx context.Context, args ...any) string {
	site := pages.CtxSite(ctx)
	if site == nil {
		return ""
	}

	q := make(url.Values)
	for i := 0; i < len(args); i += 2 {
		q.Add(fmt.Sprintf("%v", args[i]), fmt.Sprintf("%v", args[i+1]))
	}
	path := q.Get("path")
	q.Del("path")

	var link strings.Builder
	link.WriteString(strings.TrimSuffix(site.URL(), "/"))
	link.WriteRune('/')
	if path != "" && path != "/" {
		link.WriteString(strings.TrimPrefix(path, "/"))
	}

	if len(q) > 0 {
		link.WriteRune('?')
		link.WriteString(q.Encode())
	}
	return link.String()
}

func (fm *FuncMapPage) pageURLByPage(ctx context.Context, page model.Page, args ...any) string {
	if page.Site == nil {
		site := pages.CtxSite(ctx)
		if site == nil {
			return ""
		}
		page.Site = site
		page.SiteID = site.ID
	}

	path := page.URL

	if page.IsHybrid() {
		path = page.Pattern
		var rest []any
		for i := 0; i < len(args); i += 2 {
			if key := fmt.Sprintf("%v", args[i]); len(key) > 2 && key[0] == '{' && key[len(key)-1] == '}' {
				path = strings.ReplaceAll(path, key, fmt.Sprintf("%v", args[i+1]))
				continue
			}
			rest = append(rest, args[i], args[i+1])
		}
		args = rest
	}

	return fm.pageURLByPath(pages.WithSite(ctx, page.Site), append([]any{"path", path}, args...)...)
}

func (fm *FuncMapPage) findPage(ctx context.Context, id int64) model.Page {
	page, _ := fm.pageRepo.FindByID(ctx, id)
	return page
}

func (fm *FuncMapPage) pageChildren(ctx context.Context, parentID int64, now time.Time) []model.Page {
	data, _ := fm.pageRepo.FindByParentID(ctx, parentID, now)
	return data
}

func (fm *FuncMapPage) pagesByCriteria(ctx context.Context, criteria *cr.Criteria) map[string]any {
	data, total, _ := fm.pageRepo.FindAndCount(ctx, criteria)
	return map[string]any{"pages": data, "total": total}
}
