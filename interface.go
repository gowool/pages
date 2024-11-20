package pages

import "context"

type Validator interface {
	ValidateCtx(ctx context.Context, obj any) error
}

type Cache interface {
	Set(ctx context.Context, key string, value any, tags ...string) error
	Get(ctx context.Context, key string, value any) error
	DelByKey(ctx context.Context, key string) error
	DelByTag(ctx context.Context, tag string) error
}
