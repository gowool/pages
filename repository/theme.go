package repository

import (
	"context"

	"github.com/gowool/theme"
)

type ThemeRepository struct {
	r Template
}

func NewThemeRepository(r Template) theme.Repository {
	return ThemeRepository{r: r}
}

func (r ThemeRepository) FindByName(ctx context.Context, name string) (theme.Template, error) {
	return r.r.FindByName(ctx, name)
}
