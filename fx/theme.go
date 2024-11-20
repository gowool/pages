package fx

import (
	"github.com/gowool/theme"

	"github.com/gowool/pages"
	"github.com/gowool/pages/repository"
	cmstheme "github.com/gowool/pages/theme"
)

func FuncMap(pageRepo repository.Page, menu pages.Menu, matcher pages.Matcher) theme.FuncMap {
	return cmstheme.NewFuncMap(pageRepo, menu, matcher).FuncMap
}
