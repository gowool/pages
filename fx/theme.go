package fx

import (
	"github.com/gowool/theme"
	"go.uber.org/fx"

	"github.com/gowool/pages"
	"github.com/gowool/pages/repository"
	pagesheme "github.com/gowool/pages/theme"
)

func AsFuncMap(f any) any {
	return fx.Annotate(f, fx.ResultTags(`group:"theme-func-map"`))
}

func FuncMap() theme.FuncMap {
	return pagesheme.NewFuncMap().FuncMap
}

func FuncMapMenu(menu pages.Menu, matcher pages.Matcher) theme.FuncMap {
	return pagesheme.NewFuncMapMenu(menu, matcher).FuncMap
}

func FuncMapPage(pageRepo repository.Page) theme.FuncMap {
	return pagesheme.NewFuncMapPage(pageRepo).FuncMap
}
