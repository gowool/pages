package pages

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cast"
)

var _ URLGenerator = (*PageURLGenerator)(nil)

type URLGenerator interface {
	Generate(ctx context.Context, site *Site, arg any, args ...any) (string, error)
}

type PageURLGenerator struct {
	manager PageManager
}

func NewPageURLGenerator(manager PageManager) *PageURLGenerator {
	return &PageURLGenerator{
		manager: manager,
	}
}

func (g *PageURLGenerator) GenerateByPage(site *Site, page *Page, args ...any) (string, error) {
	if page == nil {
		return "", errors.New("page url generator: page is required")
	}

	if page.Site == nil {
		page.Site = site
	}

	return page.AbsURL(args...), nil
}

func (g *PageURLGenerator) GenerateByID(ctx context.Context, site *Site, id ID, args ...any) (string, error) {
	page, err := g.manager.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	return g.GenerateByPage(site, page, args...)
}

func (g *PageURLGenerator) GenerateByAlias(ctx context.Context, site *Site, alias string, args ...any) (string, error) {
	page, err := g.manager.GetByAlias(ctx, site, alias)
	if err != nil {
		return "", err
	}

	return g.GenerateByPage(site, page, args...)
}

func (g *PageURLGenerator) GenerateByURL(ctx context.Context, site *Site, url string, args ...any) (string, error) {
	page, err := g.manager.GetByURL(ctx, site, url)
	if err != nil {
		return "", err
	}

	return g.GenerateByPage(site, page, args...)
}

func (g *PageURLGenerator) GenerateByPattern(ctx context.Context, site *Site, pattern string, args ...any) (string, error) {
	page, err := g.manager.GetByPattern(ctx, site, pattern)
	if err != nil {
		return "", err
	}

	return g.GenerateByPage(site, page, args...)
}

func (g *PageURLGenerator) Generate(ctx context.Context, site *Site, arg any, args ...any) (string, error) {
	if site == nil {
		return "", errors.New("page url generator: site is required")
	}

	switch arg := arg.(type) {
	case Page:
		return g.GenerateByPage(site, &arg, args...)
	case *Page:
		return g.GenerateByPage(site, arg, args...)
	case ID:
		return g.GenerateByID(ctx, site, arg, args...)
	case string:
		if strings.HasPrefix(PageAliasPrefix, arg) {
			return g.GenerateByAlias(ctx, site, arg, args...)
		}

		if arg != "" {
			if arg == PageCMS {
				var url string

				newArgs := make([]any, 0, len(args))

				for i := 0; i < len(args); i += 2 {
					key := cast.ToString(args[i])
					switch key {
					case "url":
						url = cast.ToString(args[i+1])
					default:
						newArgs = append(newArgs, args[i], args[i+1])
					}
				}

				if url == "" {
					return "", errors.New("page url generator: url is required for cms page")
				}

				return g.GenerateByURL(ctx, site, url, newArgs...)
			}

			return g.GenerateByPattern(ctx, site, arg, args...)
		}

		var (
			pattern string
			url     string
		)

		newArgs := make([]any, 0, len(args))

		for i := 0; i < len(args); i += 2 {
			key := cast.ToString(args[i])
			switch key {
			case "pattern":
				pattern = cast.ToString(args[i+1])
			case "url":
				url = cast.ToString(args[i+1])
			default:
				newArgs = append(newArgs, args[i], args[i+1])
			}
		}

		if pattern != "" {
			if url != "" {
				newArgs = append(newArgs, "url", url)
			}
			return g.GenerateByPattern(ctx, site, pattern, newArgs...)
		}

		if url != "" {
			if pattern != "" {
				newArgs = append(newArgs, "pattern", pattern)
			}
			return g.GenerateByURL(ctx, site, url, newArgs...)
		}

		return "", errors.New("page url generator: either pattern or url is required")
	default:
		return "", fmt.Errorf("page url generator: unsupported page type %T(%v)", arg, arg)
	}
}
