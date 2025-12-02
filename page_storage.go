package pages

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/google/uuid"
)

var _ PageStorage = (*MemoryPageStorage)(nil)

type PageStorage interface {
	FindByID(ctx context.Context, id ID) (*Page, error)
	FindByURL(ctx context.Context, siteID ID, url string) (*Page, error)
	FindByPattern(ctx context.Context, siteID ID, pattern string) (*Page, error)
	FindByAlias(ctx context.Context, siteID ID, alias string) (*Page, error)
	Save(ctx context.Context, pages ...*Page) error
}

type MemoryPageStorage struct {
	data    []*Page
	ids     map[ID]int
	paths   map[string]int
	aliases map[string]int
	mu      sync.RWMutex
}

func NewMemoryPageStorage() *MemoryPageStorage {
	return &MemoryPageStorage{
		ids:     make(map[ID]int),
		paths:   make(map[string]int),
		aliases: make(map[string]int),
	}
}

func (s *MemoryPageStorage) FindByID(_ context.Context, id ID) (*Page, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if index, ok := s.ids[id]; ok {
		page := *s.data[index]
		return &page, nil
	}

	return nil, fmt.Errorf("page storage: page not found by id %s: %w", id, ErrPageNotFound)
}

func (s *MemoryPageStorage) FindByURL(_ context.Context, siteID ID, url string) (*Page, error) {
	return s.findByPath(siteID, url)
}

func (s *MemoryPageStorage) FindByPattern(_ context.Context, siteID ID, pattern string) (*Page, error) {
	return s.findByPath(siteID, pattern)
}

func (s *MemoryPageStorage) FindByAlias(_ context.Context, siteID ID, alias string) (*Page, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alias = fmt.Sprintf("%s-%s", siteID, alias)
	if index, ok := s.aliases[alias]; ok {
		page := *s.data[index]
		return &page, nil
	}

	return nil, fmt.Errorf("page storage: page not found by alias %s: %w", alias, ErrPageNotFound)
}

func (s *MemoryPageStorage) Save(_ context.Context, pages ...*Page) error {
	for _, page := range pages {
		if page.ID.IsZero() {
			page.ID = ID(uuid.NewString())
		}

		path := page.Pattern
		if page.IsCMS() {
			path = page.URL
		}
		path = fmt.Sprintf("%s-%s", page.SiteID, path)
		alias := fmt.Sprintf("%s-%s", page.SiteID, page.Alias)

		var (
			index int
			ok    bool
		)

		s.mu.Lock()

		if index, ok = s.ids[page.ID]; ok {
			s.data[index] = page
			s.deletePath(index)
			s.deleteAlias(index)
		} else {
			index = len(s.data)
			s.ids[page.ID] = index
			s.data = append(s.data, page)
		}

		if _, ok = s.paths[path]; ok {
			s.mu.Unlock()

			return fmt.Errorf("page storage: page path is not unique %s: %w", path, ErrUniqueViolation)
		}

		if _, ok = s.aliases[alias]; ok {
			s.mu.Unlock()

			return fmt.Errorf("page storage: page alias is not unique %s: %w", alias, ErrUniqueViolation)
		}

		s.paths[path] = index
		s.aliases[alias] = index

		s.mu.Unlock()
	}

	return nil
}

func (s *MemoryPageStorage) DeleteByID(_ context.Context, ids ...ID) error {
	for _, id := range ids {
		s.mu.Lock()

		if index, ok := s.ids[id]; ok {
			s.deletePath(index)
			s.deleteAlias(index)

			delete(s.ids, id)

			s.data = slices.Delete(s.data, index, index+1)

			s.mu.Unlock()
		} else {
			s.mu.Unlock()

			return fmt.Errorf("page storage: page not found by id %s: %w", id, ErrPageNotFound)
		}
	}

	return nil
}

func (s *MemoryPageStorage) deletePath(index int) {
	for k, i := range s.paths {
		if i == index {
			delete(s.paths, k)
			break
		}
	}
}

func (s *MemoryPageStorage) deleteAlias(index int) {
	for k, i := range s.aliases {
		if i == index {
			delete(s.aliases, k)
			break
		}
	}
}

func (s *MemoryPageStorage) findByPath(siteID ID, url string) (*Page, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := fmt.Sprintf("%s-%s", siteID, url)
	if index, ok := s.paths[path]; ok {
		page := *s.data[index]
		return &page, nil
	}

	return nil, fmt.Errorf("page storage: page not found by path %s: %w", path, ErrPageNotFound)
}
