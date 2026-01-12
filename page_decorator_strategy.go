package pages

import (
	"context"
	"errors"
)

var _ PageDecoratorStrategy = (*NoopPageDecoratorStrategy)(nil)

var DefaultPageDecoratorStrategy = &NoopPageDecoratorStrategy{}

type PageDecoratorStrategy interface {
	IsDecorable(ctx context.Context, pattern, uri string) (bool, error)
	IsPatternDecorable(ctx context.Context, pattern string) (bool, error)
	IsURIDecorable(ctx context.Context, uri string) (bool, error)
}

type NoopPageDecoratorStrategy struct{}

func (s *NoopPageDecoratorStrategy) IsDecorable(ctx context.Context, pattern, uri string) (bool, error) {
	ok1, err1 := s.IsPatternDecorable(ctx, pattern)
	ok2, err2 := s.IsURIDecorable(ctx, pattern)

	return ok1 && ok2, errors.Join(err1, err2)
}

func (s *NoopPageDecoratorStrategy) IsPatternDecorable(context.Context, string) (bool, error) {
	return true, nil
}

func (s *NoopPageDecoratorStrategy) IsURIDecorable(context.Context, string) (bool, error) {
	return true, nil
}
