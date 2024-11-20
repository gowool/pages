package cache

import (
	"context"
	"fmt"

	"github.com/gowool/pages"
	"github.com/gowool/pages/internal"
)

type inner[T any, ID any] interface {
	FindByID(context.Context, ID) (T, error)
	Delete(context.Context, ...ID) error
}

type repo[T any, ID any] struct {
	inner  inner[T, ID]
	cache  pages.Cache
	prefix string
}

func (r repo[T, ID]) del(ctx context.Context, id ID) {
	_ = r.cache.DelByTag(ctx, r.tag(fmt.Sprintf("%v", id)))
}

func (r repo[T, ID]) set(ctx context.Context, key string, m any, ids ...ID) {
	_ = r.cache.Set(ctx, key, m, internal.Map(ids, func(id ID) string { return r.tag(fmt.Sprintf("%v", id)) })...)
}

func (r repo[T, ID]) tag(suffix string) string {
	return fmt.Sprintf("%s:tag:%v", r.prefix, suffix)
}

func (r repo[T, ID]) findByID(ctx context.Context, id ID) (m T, err error) {
	key := fmt.Sprintf("%s:id:%v", r.prefix, id)

	if err = r.cache.Get(ctx, key, &m); err == nil {
		return
	}

	if m, err = r.inner.FindByID(ctx, id); err != nil {
		return
	}

	r.set(ctx, key, m, id)
	return
}

func (r repo[T, ID]) delete(ctx context.Context, ids ...ID) error {
	defer func() {
		for _, id := range ids {
			r.del(ctx, id)
		}
	}()

	return r.inner.Delete(ctx, ids...)
}
