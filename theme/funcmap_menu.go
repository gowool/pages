package theme

import (
	"context"
	"html/template"
	"maps"

	"github.com/gowool/theme"

	"github.com/gowool/pages"
)

type FuncMapMenu struct {
	menuService pages.Menu
	matcher     pages.Matcher
}

func NewFuncMapMenu(menu pages.Menu, matcher pages.Matcher) *FuncMapMenu {
	return &FuncMapMenu{
		menuService: menu,
		matcher:     matcher,
	}
}

func (fm *FuncMapMenu) FuncMap(t theme.Theme) template.FuncMap {
	return template.FuncMap{
		"menu":             fm.menu(t),
		"node_is_current":  fm.matcher.IsCurrent,
		"node_is_ancestor": fm.matcher.IsAncestor,
	}
}

func (fm *FuncMapMenu) menu(t theme.Theme) func(context.Context, string, string, map[string]any) template.HTML {
	return func(ctx context.Context, handle, templateName string, data map[string]any) template.HTML {
		m, err := fm.menuService.Get(ctx, handle)
		if err != nil {
			return ""
		}

		data = maps.Clone(data)
		data["node"] = m.Node

		str, err := t.HTML(ctx, templateName, data)
		if err != nil {
			return ""
		}
		return template.HTML(str)
	}
}
