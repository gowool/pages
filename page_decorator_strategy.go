package pages

import "context"

type PageDecoratorStrategy interface {
	IsPatternDecorable(ctx context.Context, pattern string) bool
	IsURIDecorable(ctx context.Context, uri string) bool
}
