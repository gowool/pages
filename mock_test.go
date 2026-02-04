package pages

import (
	"context"
	"io"
	"iter"
	"net/http"

	"github.com/stretchr/testify/mock"
)

var _ SiteStore = (*MockSiteStore)(nil)

// MockSiteStore is a mock implementation of SiteStore
type MockSiteStore struct {
	mock.Mock
}

func (m *MockSiteStore) FindPublished(ctx context.Context) iter.Seq2[*Site, error] {
	args := m.Called(ctx)
	return args.Get(0).(iter.Seq2[*Site, error])
}

var _ PageStore = (*MockPageStore)(nil)

// MockPageStore is a mock implementation of PageStore interface
type MockPageStore struct {
	mock.Mock
}

func (m *MockPageStore) FindByID(ctx context.Context, id ID) (*Page, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPageStore) FindByURL(ctx context.Context, siteID ID, url string) (*Page, error) {
	args := m.Called(ctx, siteID, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPageStore) FindByPattern(ctx context.Context, siteID ID, pattern string) (*Page, error) {
	args := m.Called(ctx, siteID, pattern)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPageStore) FindByAlias(ctx context.Context, siteID ID, alias string) (*Page, error) {
	args := m.Called(ctx, siteID, alias)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPageStore) FindByPatterns(ctx context.Context, siteID ID, patterns ...string) iter.Seq2[*Page, error] {
	args := m.Called(ctx, siteID, patterns)
	return args.Get(0).(iter.Seq2[*Page, error])
}

func (m *MockPageStore) Save(ctx context.Context, pages ...*Page) error {
	args := m.Called(ctx, pages)
	return args.Error(0)
}

var _ URLGenerator = (*MockURLGenerator)(nil)

// MockURLGenerator implements URLGenerator interface using testify/mock
type MockURLGenerator struct {
	mock.Mock
}

func (m *MockURLGenerator) Generate(ctx context.Context, site *Site, arg any, args ...any) (string, error) {
	callArgs := m.Called(ctx, site, arg, args)
	if url := callArgs.Get(0); url != nil {
		return url.(string), callArgs.Error(1)
	}
	return "", callArgs.Error(1)
}

var _ SiteRetriever = (*MockSiteRetriever)(nil)

// MockSiteRetriever is a mock implementation of SiteRetriever
type MockSiteRetriever struct {
	mock.Mock
}

func NewMockSiteRetriever(site *Site, pathInfo string, err error) *MockSiteRetriever {
	retriever := &MockSiteRetriever{}
	retriever.On("Retrieve", mock.Anything).Return(site, pathInfo, err)
	return retriever
}

func (m *MockSiteRetriever) Retrieve(r *http.Request) (*Site, string, error) {
	args := m.Called(r)
	if args.Get(0) == nil && args.Get(2) == nil {
		return nil, "", nil
	}
	if args.Get(2) != nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*Site), args.String(1), nil
}

var _ PageManager = (*MockPageManager)(nil)

// MockPageManager implements PageManager interface using testify/mock
type MockPageManager struct {
	mock.Mock
}

func (m *MockPageManager) GetByID(ctx context.Context, id ID) (*Page, error) {
	args := m.Called(ctx, id)
	if page := args.Get(0); page != nil {
		return page.(*Page), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPageManager) GetByURL(ctx context.Context, site *Site, url string) (*Page, error) {
	args := m.Called(ctx, site, url)
	if page := args.Get(0); page != nil {
		return page.(*Page), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPageManager) GetByPattern(ctx context.Context, site *Site, pattern string) (*Page, error) {
	args := m.Called(ctx, site, pattern)
	if page := args.Get(0); page != nil {
		return page.(*Page), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPageManager) GetByAlias(ctx context.Context, site *Site, alias string) (*Page, error) {
	args := m.Called(ctx, site, alias)
	if page := args.Get(0); page != nil {
		return page.(*Page), args.Error(1)
	}
	return nil, args.Error(1)
}

var _ Patterns = (*MockPatterns)(nil)

// MockPatterns is a mock implementation of Patterns interface
type MockPatterns struct {
	mock.Mock
	patterns []string
}

func NewMockPatterns(patterns []string) *MockPatterns {
	return &MockPatterns{patterns: patterns}
}

func (m *MockPatterns) Patterns() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, pattern := range m.patterns {
			if !yield(pattern) {
				return
			}
		}
	}
}

var _ PageDecoratorStrategy = (*MockPageDecoratorStrategy)(nil)

// MockPageDecoratorStrategy is a mock implementation of PageDecoratorStrategy interface
type MockPageDecoratorStrategy struct {
	mock.Mock
}

func NewMockPageDecoratorStrategy(alwaysDecorable bool) *MockPageDecoratorStrategy {
	strategy := &MockPageDecoratorStrategy{}
	if alwaysDecorable {
		strategy.On("IsPatternDecorable", mock.Anything, mock.Anything).Return(true)
		strategy.On("IsURIDecorable", mock.Anything, mock.Anything).Return(true)
	} else {
		strategy.On("IsPatternDecorable", mock.Anything, mock.Anything).Return(false)
		strategy.On("IsURIDecorable", mock.Anything, mock.Anything).Return(false)
	}
	return strategy
}

func (m *MockPageDecoratorStrategy) IsPatternDecorable(ctx context.Context, pattern string) bool {
	args := m.Called(ctx, pattern)
	return args.Bool(0)
}

func (m *MockPageDecoratorStrategy) IsURIDecorable(ctx context.Context, uri string) bool {
	args := m.Called(ctx, uri)
	return args.Bool(0)
}

var _ Theme = (*MockTheme)(nil)

type MockTheme struct {
	content  string
	template string
	err      error
	data     any
}

func (m *MockTheme) Write(_ context.Context, w io.Writer, template string, data any) error {
	m.template = template
	m.data = data
	if m.err != nil {
		return m.err
	}
	_, err := w.Write([]byte(m.content))
	return err
}

var _ PageAuthorizer = (*MockPageAuthorizer)(nil)

type MockPageAuthorizer struct {
	mock.Mock
}

func (m *MockPageAuthorizer) Authorize(ctx context.Context, action PageAction) Decision {
	args := m.Called(ctx, action)
	return args.Get(0).(Decision)
}
