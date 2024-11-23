package v1

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gowool/echox/api"
	"github.com/labstack/echo/v4"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type ConfigurationBody struct {
	Debug                 *bool                    `json:"debug,omitempty" yaml:"debug,omitempty" required:"true"`
	Multisite             *model.MultisiteStrategy `json:"multisite,omitempty" yaml:"multisite,omitempty" required:"false" enum:"host,host-by-locale,host-with-path,host-with-path-by-locale"`
	FallbackLocale        *string                  `json:"fallbackLocale,omitempty" yaml:"fallbackLocale,omitempty" required:"false"`
	IgnoreRequestPatterns *[]string                `json:"ignoreRequestPatterns,omitempty" yaml:"ignoreRequestPatterns,omitempty" required:"false"`
	IgnoreRequestURIs     *[]string                `json:"ignoreRequestURIs,omitempty" yaml:"ignoreRequestURIs,omitempty" required:"false"`
	SiteSkippers          *model.Skippers          `json:"siteSkippers,omitempty" yaml:"siteSkippers,omitempty" required:"false"`
	PageSkippers          *model.Skippers          `json:"pageSkippers,omitempty" yaml:"pageSkippers,omitempty" required:"false"`
	CatchErrors           *map[string][]int        `json:"catchErrors,omitempty" yaml:"catchErrors,omitempty" required:"false"`
	Additional            *map[string]string       `json:"additional,omitempty" yaml:"additional,omitempty" required:"false"`
}

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
		return nil, err
	}
	return &api.Response[model.Configuration]{Body: cfg}, nil
}

func (h Configuration) save(ctx context.Context, in *api.CreateInput[ConfigurationBody]) (*struct{}, error) {
	cfg, err := h.repo.Load(ctx)
	if err != nil {
		return nil, err
	}

	if in.Body.Debug != nil {
		cfg.Debug = *in.Body.Debug
	}
	if in.Body.Multisite != nil {
		cfg.Multisite = *in.Body.Multisite
	}
	if in.Body.FallbackLocale != nil {
		cfg.FallbackLocale = *in.Body.FallbackLocale
	}
	if in.Body.IgnoreRequestPatterns != nil {
		cfg.IgnoreRequestPatterns = *in.Body.IgnoreRequestPatterns
	}
	if in.Body.IgnoreRequestURIs != nil {
		cfg.IgnoreRequestURIs = *in.Body.IgnoreRequestURIs
	}
	if in.Body.SiteSkippers != nil {
		cfg.SiteSkippers = in.Body.SiteSkippers
	}
	if in.Body.PageSkippers != nil {
		cfg.PageSkippers = in.Body.PageSkippers
	}
	if in.Body.CatchErrors != nil {
		cfg.CatchErrors = *in.Body.CatchErrors
	}
	if in.Body.Additional != nil {
		cfg.Additional = *in.Body.Additional
	}

	err = h.repo.Save(ctx, &cfg)

	return nil, err
}
