package pages

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"slices"
	"sync"

	"github.com/google/uuid"
)

var _ PageStore = (*MemoryPageStore)(nil)

type PageStore interface {
	FindByID(ctx context.Context, id ID) (*Page, error)
	FindByURL(ctx context.Context, siteID ID, url string) (*Page, error)
	FindByPattern(ctx context.Context, siteID ID, pattern string) (*Page, error)
	FindByPatterns(ctx context.Context, siteID ID, patterns ...string) iter.Seq2[*Page, error]
	FindByAlias(ctx context.Context, siteID ID, alias string) (*Page, error)
	Save(ctx context.Context, pages ...*Page) error
}

type MemoryPageStore struct {
	data    []*Page
	ids     map[ID]int
	paths   map[string]int
	aliases map[string]int
	mu      sync.RWMutex
}

func NewMemoryPageStore() *MemoryPageStore {
	return &MemoryPageStore{
		ids:     make(map[ID]int),
		paths:   make(map[string]int),
		aliases: make(map[string]int),
	}
}

func (s *MemoryPageStore) FindByID(_ context.Context, id ID) (*Page, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if index, ok := s.ids[id]; ok {
		return s.data[index].Copy(), nil
	}

	return nil, fmt.Errorf("page store: page not found by id %s: %w", id, ErrPageNotFound)
}

func (s *MemoryPageStore) FindByURL(_ context.Context, siteID ID, url string) (*Page, error) {
	return s.findByPath(siteID, url)
}

func (s *MemoryPageStore) FindByPattern(_ context.Context, siteID ID, pattern string) (*Page, error) {
	return s.findByPath(siteID, pattern)
}

func (s *MemoryPageStore) FindByPatterns(_ context.Context, siteID ID, patterns ...string) iter.Seq2[*Page, error] {
	return func(yield func(*Page, error) bool) {
		if len(patterns) == 0 {
			return
		}

		visited := make(map[string]struct{})
		for _, pattern := range patterns {
			if _, ok := visited[pattern]; ok {
				continue
			}

			visited[pattern] = struct{}{}

			page, err := s.findByPath(siteID, pattern)
			if errors.Is(err, ErrPageNotFound) {
				continue
			}

			if !yield(page, err) {
				return
			}
		}
	}
}

func (s *MemoryPageStore) FindByAlias(_ context.Context, siteID ID, alias string) (*Page, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alias = fmt.Sprintf("%s-%s", siteID, alias)
	if index, ok := s.aliases[alias]; ok {
		return s.data[index].Copy(), nil
	}

	return nil, fmt.Errorf("page store: page not found by alias %s: %w", alias, ErrPageNotFound)
}

func (s *MemoryPageStore) Save(_ context.Context, pages ...*Page) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

		if index, ok = s.paths[path]; ok && s.data[index].ID != page.ID {
			return fmt.Errorf("page store: page path is not unique %s: %w", path, ErrUniqueViolation)
		}

		// Only check alias uniqueness if alias is not empty
		if page.Alias != "" {
			if index, ok = s.aliases[alias]; ok && s.data[index].ID != page.ID {
				return fmt.Errorf("page store: page alias is not unique %s: %w", alias, ErrUniqueViolation)
			}
		}

		if index, ok = s.ids[page.ID]; ok {
			s.data[index] = page
			s.deletePath(index)
			s.deleteAlias(index)
		} else {
			index = len(s.data)
			s.ids[page.ID] = index
			// Make the append operation as atomic as possible
			newData := append(s.data, page)
			s.data = newData
		}

		s.paths[path] = index

		if page.Alias != "" {
			s.aliases[alias] = index
		}
	}

	return nil
}

func (s *MemoryPageStore) DeleteByID(_ context.Context, ids ...ID) error {
	for _, id := range ids {
		s.mu.Lock()

		if index, ok := s.ids[id]; ok {
			s.deletePath(index)
			s.deleteAlias(index)

			delete(s.ids, id)

			s.data = slices.Delete(s.data, index, index+1)

			// Update all indices after the deleted index
			s.updateIndices(index)

			s.mu.Unlock()
		} else {
			s.mu.Unlock()

			return fmt.Errorf("page store: page not found by id %s: %w", id, ErrPageNotFound)
		}
	}

	return nil
}

func (s *MemoryPageStore) deletePath(index int) {
	for k, i := range s.paths {
		if i == index {
			delete(s.paths, k)
			break
		}
	}
}

func (s *MemoryPageStore) deleteAlias(index int) {
	for k, i := range s.aliases {
		if i == index {
			delete(s.aliases, k)
			break
		}
	}
}

func (s *MemoryPageStore) updateIndices(deletedIndex int) {
	// Update indices in ids map
	for id, idx := range s.ids {
		if idx > deletedIndex {
			s.ids[id] = idx - 1
		}
	}

	// Update indices in paths map
	for path, idx := range s.paths {
		if idx > deletedIndex {
			s.paths[path] = idx - 1
		}
	}

	// Update indices in aliases map
	for alias, idx := range s.aliases {
		if idx > deletedIndex {
			s.aliases[alias] = idx - 1
		}
	}
}

func (s *MemoryPageStore) findByPath(siteID ID, url string) (*Page, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := fmt.Sprintf("%s-%s", siteID, url)
	if index, ok := s.paths[path]; ok {
		return s.data[index].Copy(), nil
	}

	return nil, fmt.Errorf("page store: page not found by path %s: %w", path, ErrPageNotFound)
}

// GetData returns a copy of the data slice for testing purposes.
// This method provides thread-safe access to the internal data.
func (s *MemoryPageStore) GetData() []*Page {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Page, len(s.data))
	copy(result, s.data)
	return result
}
