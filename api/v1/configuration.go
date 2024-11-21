package v1

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gowool/echox/api"
	"github.com/labstack/echo/v4"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type Configuration struct {
	errorTransformer api.ErrorTransformerFunc
	repo             repository.Configuration
	op               func(options ...api.Option) huma.Operation
}

func NewConfiguration(repo repository.Configuration, errorTransformer api.ErrorTransformerFunc, options ...api.Option) Configuration {
	opts := make([]api.Option, 0, len(options)+2)
	opts = append(opts, options...)
	opts = append(opts, api.WithPath("/pages/configuration"), api.WithAddTags("page"))

	return Configuration{
		errorTransformer: errorTransformer,
		repo:             repo,
		op:               api.Operation(opts...),
	}
}

func (Configuration) Area() string {
	return Info.Area
}

func (Configuration) Version() string {
	return Info.Version
}

func (h Configuration) Register(_ *echo.Echo, humaAPI huma.API) {
	api.Register(humaAPI, api.Transform(h.errorTransformer, h.load), h.op(api.WithSummary("Get configuration")))
	api.Register(humaAPI, api.Transform(h.errorTransformer, h.save), h.op(api.WithPatch, api.WithSummary("Save configuration")))
}

func (h Configuration) load(ctx context.Context, _ *struct{}) (*api.Response[model.Configuration], error) {
	cfg, err := h.repo.Load(ctx)
	if err != nil {
		return nil, h.errorTransformer(ctx, err)
	}
	return &api.Response[model.Configuration]{Body: cfg}, nil
}

func (h Configuration) save(ctx context.Context, in *api.CreateInput[model.Configuration]) (*struct{}, error) {
	cfg, err := h.repo.Load(ctx)
	if err != nil {
		return nil, h.errorTransformer(ctx, err)
	}

	cfg = cfg.With(in.Body)
	if err = h.repo.Save(ctx, &cfg); err != nil {
		return nil, h.errorTransformer(ctx, err)
	}
	return nil, nil
}
