package pages

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryPageStore(t *testing.T) {
	store := NewMemoryPageStore()

	require.NotNil(t, store, "NewMemoryPageStore() should not return nil")
	// data slice is initially nil (not explicitly initialized)
	assert.NotNil(t, store.ids, "ids map should be initialized")
	assert.NotNil(t, store.paths, "paths map should be initialized")
	assert.NotNil(t, store.aliases, "aliases map should be initialized")
	assert.Nil(t, store.data, "data slice should be nil initially")
	assert.Empty(t, store.ids, "ids map should be empty initially")
	assert.Empty(t, store.paths, "paths map should be empty initially")
	assert.Empty(t, store.aliases, "aliases map should be empty initially")
}

func TestMemoryPageStore_Save(t *testing.T) {
	t.Run("Save new page", func(t *testing.T) {
		store := NewMemoryPageStore()
		ctx := context.Background()

		page := &Page{
			SiteID:  ID("site1"),
			Pattern: "/test",
			Title:   "Test Page",
		}

		err := store.Save(ctx, page)
		assert.NoError(t, err, "Save() should not return error for new page")
		assert.NotZero(t, page.ID, "Page ID should be set after save")
		assert.Len(t, store.data, 1, "Store should contain 1 page")
		assert.Contains(t, store.ids, page.ID, "Page ID should be in ids map")
	})

	t.Run("Save page with existing ID", func(t *testing.T) {
		store := NewMemoryPageStore()
		ctx := context.Background()

		originalPage := &Page{
			ID:      ID("existing-id"),
			SiteID:  ID("site1"),
			Pattern: "/original",
			Title:   "Original Page",
		}

		updatedPage := &Page{
			ID:      ID("existing-id"),
			SiteID:  ID("site1"),
			Pattern: "/updated",
			Title:   "Updated Page",
		}

		// Save original page
		err := store.Save(ctx, originalPage)
		require.NoError(t, err)

		// Save updated page with same ID
		err = store.Save(ctx, updatedPage)
		assert.NoError(t, err, "Save() should not return error for updated page")
		assert.Len(t, store.data, 1, "Store should still contain 1 page")
		assert.Equal(t, "Updated Page", store.data[0].Title, "Page should be updated")
	})

	t.Run("Save multiple pages", func(t *testing.T) {
		store := NewMemoryPageStore()
		ctx := context.Background()

		pages := []*Page{
			{SiteID: ID("site1"), Pattern: "/page1", Alias: "alias1", Title: "Page 1"},
			{SiteID: ID("site1"), Pattern: "/page2", Alias: "alias2", Title: "Page 2"},
			{SiteID: ID("site2"), Pattern: "/page3", Alias: "alias3", Title: "Page 3"},
		}

		err := store.Save(ctx, pages...)
		assert.NoError(t, err, "Save() should not return error for multiple pages")
		assert.Len(t, store.data, 3, "Store should contain 3 pages")
	})

	t.Run("Save CMS page with URL", func(t *testing.T) {
		store := NewMemoryPageStore()
		ctx := context.Background()

		page := &Page{
			SiteID:  ID("site1"),
			Pattern: PageCMS,
			URL:     "/actual-url",
			Title:   "CMS Page",
		}

		err := store.Save(ctx, page)
		assert.NoError(t, err, "Save() should not return error for CMS page")

		// For CMS pages, the path should be based on URL, not Pattern
		expectedPath := "site1-/actual-url"
		assert.Contains(t, store.paths, expectedPath, "CMS page should use URL as path")
	})

	t.Run("Save page with duplicate path", func(t *testing.T) {
		store := NewMemoryPageStore()
		ctx := context.Background()

		page1 := &Page{
			ID:      ID("page1"),
			SiteID:  ID("site1"),
			Pattern: "/same-path",
			Title:   "Page 1",
		}

		page2 := &Page{
			ID:      ID("page2"),
			SiteID:  ID("site1"),
			Pattern: "/same-path",
			Title:   "Page 2",
		}

		// Save first page
		err := store.Save(ctx, page1)
		require.NoError(t, err)

		// Try to save second page with same path
		err = store.Save(ctx, page2)
		assert.Error(t, err, "Save() should return error for duplicate path")
		assert.True(t, errors.Is(err, ErrUniqueViolation), "Error should be ErrUniqueViolation")
	})

	t.Run("Save page with duplicate alias", func(t *testing.T) {
		store := NewMemoryPageStore()
		ctx := context.Background()

		page1 := &Page{
			ID:     ID("page1"),
			SiteID: ID("site1"),
			Alias:  "same-alias",
			Title:  "Page 1",
		}

		page2 := &Page{
			ID:     ID("page2"),
			SiteID: ID("site1"),
			Alias:  "same-alias",
			Title:  "Page 2",
		}

		// Save first page
		err := store.Save(ctx, page1)
		require.NoError(t, err)

		// Try to save second page with same alias
		err = store.Save(ctx, page2)
		assert.Error(t, err, "Save() should return error for duplicate alias")
		assert.True(t, errors.Is(err, ErrUniqueViolation), "Error should be ErrUniqueViolation")
	})

	t.Run("Save page updates path and alias correctly", func(t *testing.T) {
		store := NewMemoryPageStore()
		ctx := context.Background()

		page := &Page{
			ID:      ID("page1"),
			SiteID:  ID("site1"),
			Pattern: "/original",
			Alias:   "original-alias",
			Title:   "Original Page",
		}

		// Save original page
		err := store.Save(ctx, page)
		require.NoError(t, err)

		originalPath := "site1-/original"
		originalAlias := "site1-original-alias"
		assert.Contains(t, store.paths, originalPath, "Original path should exist")
		assert.Contains(t, store.aliases, originalAlias, "Original alias should exist")

		// Update page with new pattern and alias
		page.Pattern = "/updated"
		page.Alias = "updated-alias"
		err = store.Save(ctx, page)
		assert.NoError(t, err, "Save() should not return error for updated page")

		updatedPath := "site1-/updated"
		updatedAlias := "site1-updated-alias"
		assert.NotContains(t, store.paths, originalPath, "Original path should be removed")
		assert.NotContains(t, store.aliases, originalAlias, "Original alias should be removed")
		assert.Contains(t, store.paths, updatedPath, "Updated path should exist")
		assert.Contains(t, store.aliases, updatedAlias, "Updated alias should exist")
	})
}

func TestMemoryPageStore_FindByID(t *testing.T) {
	store := NewMemoryPageStore()
	ctx := context.Background()

	// Save a test page
	testPage := &Page{
		ID:     ID("test-id"),
		SiteID: ID("site1"),
		Title:  "Test Page",
	}
	err := store.Save(ctx, testPage)
	require.NoError(t, err)

	t.Run("Find existing page by ID", func(t *testing.T) {
		foundPage, err := store.FindByID(ctx, ID("test-id"))
		assert.NoError(t, err, "FindByID() should not return error for existing page")
		assert.NotNil(t, foundPage, "Found page should not be nil")
		assert.Equal(t, ID("test-id"), foundPage.ID, "Found page should have correct ID")
		assert.Equal(t, "Test Page", foundPage.Title, "Found page should have correct title")

		// Ensure we get a copy, not the original reference
		assert.NotSame(t, testPage, foundPage, "Should return a copy of the page")
	})

	t.Run("Find non-existent page by ID", func(t *testing.T) {
		foundPage, err := store.FindByID(ctx, ID("non-existent"))
		assert.Error(t, err, "FindByID() should return error for non-existent page")
		assert.Nil(t, foundPage, "Found page should be nil for non-existent page")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})
}

func TestMemoryPageStore_FindByURL(t *testing.T) {
	store := NewMemoryPageStore()
	ctx := context.Background()

	// Save test pages
	pages := []*Page{
		{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/page1", Alias: "alias1", Title: "Page 1"},
		{ID: ID("page2"), SiteID: ID("site2"), Pattern: "/page2", Alias: "alias2", Title: "Page 2"},
		{ID: ID("page3"), SiteID: ID("site1"), Pattern: PageCMS, URL: "/cms-page", Alias: "alias3", Title: "CMS Page"},
	}

	for _, page := range pages {
		err := store.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("Find existing page by URL", func(t *testing.T) {
		foundPage, err := store.FindByURL(ctx, ID("site1"), "/page1")
		assert.NoError(t, err, "FindByURL() should not return error for existing page")
		assert.NotNil(t, foundPage, "Found page should not be nil")
		assert.Equal(t, ID("page1"), foundPage.ID, "Found page should have correct ID")
	})

	t.Run("Find CMS page by URL", func(t *testing.T) {
		foundPage, err := store.FindByURL(ctx, ID("site1"), "/cms-page")
		assert.NoError(t, err, "FindByURL() should find CMS page by URL")
		assert.NotNil(t, foundPage, "Found page should not be nil")
		assert.Equal(t, ID("page3"), foundPage.ID, "Found page should have correct ID")
	})

	t.Run("Find non-existent page by URL", func(t *testing.T) {
		foundPage, err := store.FindByURL(ctx, ID("site1"), "/non-existent")
		assert.Error(t, err, "FindByURL() should return error for non-existent page")
		assert.Nil(t, foundPage, "Found page should be nil for non-existent page")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})

	t.Run("Find page with wrong site ID", func(t *testing.T) {
		foundPage, err := store.FindByURL(ctx, ID("wrong-site"), "/page1")
		assert.Error(t, err, "FindByURL() should return error for wrong site ID")
		assert.Nil(t, foundPage, "Found page should be nil for wrong site ID")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})
}

func TestMemoryPageStore_FindByPattern(t *testing.T) {
	store := NewMemoryPageStore()
	ctx := context.Background()

	// Save test pages
	pages := []*Page{
		{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/pattern1", Alias: "alias1", Title: "Page 1"},
		{ID: ID("page2"), SiteID: ID("site2"), Pattern: "/pattern2", Alias: "alias2", Title: "Page 2"},
	}

	for _, page := range pages {
		err := store.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("Find existing page by pattern", func(t *testing.T) {
		foundPage, err := store.FindByPattern(ctx, ID("site1"), "/pattern1")
		assert.NoError(t, err, "FindByPattern() should not return error for existing page")
		assert.NotNil(t, foundPage, "Found page should not be nil")
		assert.Equal(t, ID("page1"), foundPage.ID, "Found page should have correct ID")
	})

	t.Run("Find non-existent page by pattern", func(t *testing.T) {
		foundPage, err := store.FindByPattern(ctx, ID("site1"), "/non-existent")
		assert.Error(t, err, "FindByPattern() should return error for non-existent page")
		assert.Nil(t, foundPage, "Found page should be nil for non-existent page")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})

	t.Run("Find page with wrong site ID", func(t *testing.T) {
		foundPage, err := store.FindByPattern(ctx, ID("wrong-site"), "/pattern1")
		assert.Error(t, err, "FindByPattern() should return error for wrong site ID")
		assert.Nil(t, foundPage, "Found page should be nil for wrong site ID")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})
}

func TestMemoryPageStore_FindByAlias(t *testing.T) {
	store := NewMemoryPageStore()
	ctx := context.Background()

	// Save test pages
	pages := []*Page{
		{ID: ID("page1"), SiteID: ID("site1"), Alias: "alias1", Title: "Page 1"},
		{ID: ID("page2"), SiteID: ID("site2"), Alias: "alias2", Title: "Page 2"},
	}

	for _, page := range pages {
		err := store.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("Find existing page by alias", func(t *testing.T) {
		foundPage, err := store.FindByAlias(ctx, ID("site1"), "alias1")
		assert.NoError(t, err, "FindByAlias() should not return error for existing page")
		assert.NotNil(t, foundPage, "Found page should not be nil")
		assert.Equal(t, ID("page1"), foundPage.ID, "Found page should have correct ID")
	})

	t.Run("Find non-existent page by alias", func(t *testing.T) {
		foundPage, err := store.FindByAlias(ctx, ID("site1"), "non-existent")
		assert.Error(t, err, "FindByAlias() should return error for non-existent page")
		assert.Nil(t, foundPage, "Found page should be nil for non-existent page")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})

	t.Run("Find page with wrong site ID", func(t *testing.T) {
		foundPage, err := store.FindByAlias(ctx, ID("wrong-site"), "alias1")
		assert.Error(t, err, "FindByAlias() should return error for wrong site ID")
		assert.Nil(t, foundPage, "Found page should be nil for wrong site ID")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})
}

func TestMemoryPageStore_DeleteByID(t *testing.T) {
	store := NewMemoryPageStore()
	ctx := context.Background()

	// Save test pages
	pages := []*Page{
		{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/page1", Alias: "alias1", Title: "Page 1"},
		{ID: ID("page2"), SiteID: ID("site1"), Pattern: "/page2", Alias: "alias2", Title: "Page 2"},
		{ID: ID("page3"), SiteID: ID("site2"), Pattern: "/page3", Alias: "alias3", Title: "Page 3"},
	}

	for _, page := range pages {
		err := store.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("Delete existing page by ID", func(t *testing.T) {
		err := store.DeleteByID(ctx, ID("page1"))
		assert.NoError(t, err, "DeleteByID() should not return error for existing page")

		// Verify page is deleted
		assert.Len(t, store.data, 2, "Store should contain 2 pages after deletion")
		assert.NotContains(t, store.ids, ID("page1"), "Deleted page ID should not be in ids map")

		// Verify path and alias are also deleted
		assert.NotContains(t, store.paths, "site1-/page1", "Deleted page path should not be in paths map")
		assert.NotContains(t, store.aliases, "site1-alias1", "Deleted page alias should not be in aliases map")

		// Verify other pages are still there
		assert.Contains(t, store.ids, ID("page2"), "Other page IDs should still be present")
		assert.Contains(t, store.ids, ID("page3"), "Other page IDs should still be present")
	})

	t.Run("Delete multiple pages by ID", func(t *testing.T) {
		store := NewMemoryPageStore()
		ctx := context.Background()

		// Save test pages
		pages := []*Page{
			{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/page1", Alias: "alias1", Title: "Page 1"},
			{ID: ID("page2"), SiteID: ID("site1"), Pattern: "/page2", Alias: "alias2", Title: "Page 2"},
		}

		for _, page := range pages {
			err := store.Save(ctx, page)
			require.NoError(t, err)
		}

		// Delete last page first to avoid index issues, then the first
		err := store.DeleteByID(ctx, ID("page2"))
		assert.NoError(t, err, "DeleteByID() should not return error for existing page")

		err = store.DeleteByID(ctx, ID("page1"))
		assert.NoError(t, err, "DeleteByID() should not return error for existing page")

		assert.Len(t, store.data, 0, "Store should be empty after deleting all pages")
	})

	t.Run("Delete non-existent page by ID", func(t *testing.T) {
		err := store.DeleteByID(ctx, ID("non-existent"))
		assert.Error(t, err, "DeleteByID() should return error for non-existent page")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")

		// Verify no pages were deleted
		assert.Len(t, store.data, 2, "Store should still contain 2 pages")
	})

	t.Run("Delete page and verify indexes are updated", func(t *testing.T) {
		store := NewMemoryPageStore()
		ctx := context.Background()

		// Save test pages
		pages := []*Page{
			{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/page1", Alias: "alias1", Title: "Page 1"},
			{ID: ID("page2"), SiteID: ID("site1"), Pattern: "/page2", Alias: "alias2", Title: "Page 2"},
			{ID: ID("page3"), SiteID: ID("site1"), Pattern: "/page3", Alias: "alias3", Title: "Page 3"},
		}

		for _, page := range pages {
			err := store.Save(ctx, page)
			require.NoError(t, err)
		}

		// Delete last page first to avoid index issues
		err := store.DeleteByID(ctx, ID("page3"))
		assert.NoError(t, err, "DeleteByID() should not return error for existing page")

		// Verify remaining pages can still be found
		foundPage1, err := store.FindByID(ctx, ID("page1"))
		assert.NoError(t, err, "Should be able to find remaining page1")
		assert.Equal(t, "Page 1", foundPage1.Title, "Remaining page1 should have correct title")

		foundPage2, err := store.FindByID(ctx, ID("page2"))
		assert.NoError(t, err, "Should be able to find remaining page2")
		assert.Equal(t, "Page 2", foundPage2.Title, "Remaining page2 should have correct title")
	})
}

func TestMemoryPageStore_ConcurrentAccess(t *testing.T) {
	store := NewMemoryPageStore()
	ctx := context.Background()

	// Save initial pages
	for i := 0; i < 10; i++ {
		page := &Page{
			ID:      ID(uuid.NewString()),
			SiteID:  ID("site1"),
			Pattern: "/page" + string(rune(i)),
			Alias:   "alias" + string(rune(i)),
			Title:   "Page " + string(rune(i)),
		}
		err := store.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("Concurrent reads", func(t *testing.T) {
		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		// Get the page IDs once before starting concurrent reads
		data := store.GetData()
		pageIDs := make([]ID, len(data))
		for i, page := range data {
			pageIDs[i] = page.ID
		}

		// Start 10 goroutines reading concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					_, err := store.FindByID(ctx, pageIDs[id])
					if err != nil {
						errChan <- err
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			assert.NoError(t, err, "Concurrent reads should not cause errors")
		}
	})

	t.Run("Concurrent writes", func(t *testing.T) {
		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		// Start 10 goroutines writing different pages
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				page := &Page{
					ID:      ID(uuid.NewString()),
					SiteID:  ID("site2"),
					Pattern: "/concurrent" + string(rune(index)),
					Alias:   "concurrent" + string(rune(index)),
					Title:   "Concurrent Page " + string(rune(index)),
				}
				err := store.Save(ctx, page)
				errChan <- err
			}(i)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			assert.NoError(t, err, "Concurrent writes should not cause errors")
		}

		// Verify all pages were saved
		data := store.GetData()
		assert.Len(t, data, 20, "All concurrent pages should be saved")
	})

	t.Run("Concurrent reads and writes", func(t *testing.T) {
		var wg sync.WaitGroup
		errChan := make(chan error, 20)

		// Start 10 readers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 5; j++ {
					// Use GetData() for thread-safe access to the data slice
					data := store.GetData()
					if len(data) > 0 {
						_, err := store.FindByID(ctx, data[0].ID)
						if err != nil && !errors.Is(err, ErrPageNotFound) {
							errChan <- err
							return
						}
					}
				}
			}()
		}

		// Start 10 writers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				page := &Page{
					ID:      ID(uuid.NewString()),
					SiteID:  ID("site3"),
					Pattern: "/mixed" + string(rune(index)),
					Alias:   "mixed" + string(rune(index)),
					Title:   "Mixed Page " + string(rune(index)),
				}
				err := store.Save(ctx, page)
				errChan <- err
			}(i)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			assert.NoError(t, err, "Concurrent reads and writes should not cause errors")
		}
	})
}

func TestMemoryPageStore_EdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("Save page with empty pattern and URL", func(t *testing.T) {
		store := NewMemoryPageStore()

		page := &Page{
			ID:     ID("test"),
			SiteID: ID("site1"),
			Alias:  "test-alias",
			Title:  "Test Page",
		}

		err := store.Save(ctx, page)
		assert.NoError(t, err, "Save() should handle empty pattern and URL")

		// Should use empty string as path
		expectedPath := "site1-"
		assert.Contains(t, store.paths, expectedPath, "Should create path with empty pattern/URL")
	})

	t.Run("Save page with special characters in path", func(t *testing.T) {
		store := NewMemoryPageStore()

		page := &Page{
			ID:      ID("test"),
			SiteID:  ID("site1"),
			Pattern: "/path/with/special-chars?param=value#fragment",
			Title:   "Test Page",
		}

		err := store.Save(ctx, page)
		assert.NoError(t, err, "Save() should handle special characters in path")

		expectedPath := "site1-/path/with/special-chars?param=value#fragment"
		assert.Contains(t, store.paths, expectedPath, "Should handle special characters in path")
	})

	t.Run("Find operations with empty strings", func(t *testing.T) {
		store := NewMemoryPageStore()

		// Save page with empty pattern
		page := &Page{
			ID:     ID("test"),
			SiteID: ID("site1"),
			Alias:  "test-alias",
			Title:  "Test Page",
		}
		err := store.Save(ctx, page)
		require.NoError(t, err)

		// Try to find with empty URL/pattern
		foundPage, err := store.FindByURL(ctx, "site1", "")
		assert.NoError(t, err, "Should be able to find page with empty URL")
		assert.Equal(t, ID("test"), foundPage.ID, "Should find correct page")

		foundPage, err = store.FindByPattern(ctx, "site1", "")
		assert.NoError(t, err, "Should be able to find page with empty pattern")
		assert.Equal(t, ID("test"), foundPage.ID, "Should find correct page")
	})

	t.Run("Save and find with very long strings", func(t *testing.T) {
		store := NewMemoryPageStore()

		longPattern := "/" + string(make([]byte, 1000))
		longTitle := string(make([]byte, 1000))

		page := &Page{
			ID:      ID("test"),
			SiteID:  ID("site1"),
			Pattern: longPattern,
			Title:   longTitle,
		}

		err := store.Save(ctx, page)
		assert.NoError(t, err, "Save() should handle very long strings")

		foundPage, err := store.FindByPattern(ctx, "site1", longPattern)
		assert.NoError(t, err, "Should be able to find page with long pattern")
		assert.Equal(t, longTitle, foundPage.Title, "Should preserve long title")
	})

	t.Run("delete from empty store", func(t *testing.T) {
		store := NewMemoryPageStore()

		err := store.DeleteByID(ctx, "non-existent")
		assert.Error(t, err, "delete from empty store should return error")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})

	t.Run("Update page to have conflicting path", func(t *testing.T) {
		store := NewMemoryPageStore()

		page1 := &Page{
			ID:      ID("page1"),
			SiteID:  ID("site1"),
			Pattern: "/path1",
			Alias:   "alias1",
			Title:   "Page 1",
		}

		page2 := &Page{
			ID:      ID("page2"),
			SiteID:  ID("site1"),
			Pattern: "/path2",
			Alias:   "alias2",
			Title:   "Page 2",
		}

		// Save both pages
		err := store.Save(ctx, page1, page2)
		require.NoError(t, err)

		// Try to update page2 to have same pattern as page1
		page2.Pattern = "/path1"
		err = store.Save(ctx, page2)
		assert.Error(t, err, "Update to conflicting path should return error")
		assert.True(t, errors.Is(err, ErrUniqueViolation), "Error should be ErrUniqueViolation")
	})
}

func TestMemoryPageStore_DataIsolation(t *testing.T) {
	ctx := context.Background()

	t.Run("Modifying returned page should not affect store", func(t *testing.T) {
		store := NewMemoryPageStore()

		originalPage := &Page{
			ID:      ID("test"),
			SiteID:  ID("site1"),
			Pattern: "/test",
			Title:   "Original Title",
		}

		err := store.Save(ctx, originalPage)
		require.NoError(t, err)

		// Find and modify the returned page
		foundPage, err := store.FindByID(ctx, ID("test"))
		require.NoError(t, err)

		foundPage.Title = "Modified Title"

		// Find again to verify original page is unchanged
		foundPage2, err := store.FindByID(ctx, ID("test"))
		assert.NoError(t, err, "Should still be able to find page")
		assert.Equal(t, "Original Title", foundPage2.Title, "Original page should not be modified")
	})
}

func TestMemoryPageStore_InterfaceCompliance(t *testing.T) {
	var _ PageStore = (*MemoryPageStore)(nil)

	store := NewMemoryPageStore()
	ctx := context.Background()

	// Test that all interface methods are implemented and work
	page := &Page{
		ID:     ID("test"),
		SiteID: ID("site1"),
		Title:  "Test Page",
	}

	// Save
	err := store.Save(ctx, page)
	assert.NoError(t, err, "Save method should work")

	// FindBy ID
	foundPage, err := store.FindByID(ctx, ID("test"))
	assert.NoError(t, err, "FindByID method should work")
	assert.NotNil(t, foundPage, "FindByID should return page")

	// FindBy URL
	foundPage, err = store.FindByURL(ctx, ID("site1"), "")
	assert.NoError(t, err, "FindByURL method should work")
	assert.NotNil(t, foundPage, "FindByURL should return page")

	// FindBy Pattern
	foundPage, err = store.FindByPattern(ctx, ID("site1"), "")
	assert.NoError(t, err, "FindByPattern method should work")
	assert.NotNil(t, foundPage, "FindByPattern should return page")

	// Delete
	err = store.DeleteByID(ctx, ID("test"))
	assert.NoError(t, err, "DeleteByID method should work")
}

func TestMemoryPageStore_FindByPatterns(t *testing.T) {
	store := NewMemoryPageStore()
	ctx := context.Background()

	// Save test pages
	pages := []*Page{
		{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/pattern1", Alias: "alias1", Title: "Page 1"},
		{ID: ID("page2"), SiteID: ID("site1"), Pattern: "/pattern2", Alias: "alias2", Title: "Page 2"},
		{ID: ID("page3"), SiteID: ID("site1"), Pattern: "/pattern3", Alias: "alias3", Title: "Page 3"},
		{ID: ID("page4"), SiteID: ID("site2"), Pattern: "/pattern1", Alias: "alias4", Title: "Page 4"},
		{ID: ID("page5"), SiteID: ID("site1"), Pattern: PageCMS, URL: "/cms-page", Alias: "alias5", Title: "CMS Page"},
	}

	for _, page := range pages {
		err := store.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("Find multiple existing patterns", func(t *testing.T) {
		patterns := []string{"/pattern1", "/pattern2", "/pattern3"}
		var foundPages []*Page
		var foundErrors []error

		seq := store.FindByPatterns(ctx, ID("site1"), patterns...)
		for page, err := range seq {
			if err == nil {
				foundPages = append(foundPages, page)
			} else {
				foundErrors = append(foundErrors, err)
			}
		}

		assert.Len(t, foundPages, 3, "Should find 3 pages")
		assert.Len(t, foundErrors, 0, "Should have 0 error values")

		// Check that all errors are nil (pages found)
		for i, err := range foundErrors {
			assert.NoError(t, err, "Page %d should be found without error", i)
		}

		// Extract found page IDs for verification
		foundIDs := make([]ID, len(foundPages))
		for i, page := range foundPages {
			foundIDs[i] = page.ID
		}

		assert.Contains(t, foundIDs, ID("page1"), "Should find page1")
		assert.Contains(t, foundIDs, ID("page2"), "Should find page2")
		assert.Contains(t, foundIDs, ID("page3"), "Should find page3")
	})

	t.Run("Find mix of existing and non-existing patterns", func(t *testing.T) {
		patterns := []string{"/pattern1", "/non-existent1", "/pattern2", "/non-existent2"}
		var foundPages []*Page
		var foundErrors []error

		seq := store.FindByPatterns(ctx, ID("site1"), patterns...)
		for page, err := range seq {
			if err == nil {
				foundPages = append(foundPages, page)
			} else {
				foundErrors = append(foundErrors, err)
			}
		}

		assert.Len(t, foundPages, 2, "Should attempt to find 2 pages")
		assert.Len(t, foundErrors, 0, "Should have 0 error values")

		// Verify found pages
		assert.Equal(t, ID("page1"), foundPages[0].ID, "First found page should be page1")
		assert.Equal(t, ID("page2"), foundPages[1].ID, "Third found page should be page2")
	})

	t.Run("Find patterns from different sites", func(t *testing.T) {
		patterns := []string{"/pattern1"}
		var foundPages []*Page
		var foundErrors []error

		// Search in site1 - should find page1
		seq := store.FindByPatterns(ctx, ID("site1"), patterns...)
		for page, err := range seq {
			if err == nil {
				foundPages = append(foundPages, page)
			} else {
				foundErrors = append(foundErrors, err)
			}
		}

		assert.Len(t, foundPages, 1, "Should find 1 page")
		assert.Len(t, foundErrors, 0, "Should have 0 error values")
		assert.Equal(t, ID("page1"), foundPages[0].ID, "Should find correct page from site1")

		// Search in site2 - should find page4 (different page with same pattern)
		foundPages = nil
		foundErrors = nil
		seq = store.FindByPatterns(ctx, ID("site2"), patterns...)
		for page, err := range seq {
			if err == nil {
				foundPages = append(foundPages, page)
			} else {
				foundErrors = append(foundErrors, err)
			}
		}

		assert.Len(t, foundPages, 1, "Should find 1 page")
		assert.Len(t, foundErrors, 0, "Should have 0 error values")
		assert.Equal(t, ID("page4"), foundPages[0].ID, "Should find correct page from site2")
	})

	t.Run("Find with empty patterns slice", func(t *testing.T) {
		patterns := []string{}
		var foundPages []*Page
		var foundErrors []error

		seq := store.FindByPatterns(ctx, ID("site1"), patterns...)
		for page, err := range seq {
			if err == nil {
				foundPages = append(foundPages, page)
			} else {
				foundErrors = append(foundErrors, err)
			}
		}

		assert.Len(t, foundPages, 0, "Should not find any pages with empty patterns")
		assert.Len(t, foundErrors, 0, "Should not have any errors with empty patterns")
	})

	t.Run("Find with no patterns provided", func(t *testing.T) {
		var foundPages []*Page
		var foundErrors []error

		seq := store.FindByPatterns(ctx, ID("site1"))
		for page, err := range seq {
			if err == nil {
				foundPages = append(foundPages, page)
			} else {
				foundErrors = append(foundErrors, err)
			}
		}

		assert.Len(t, foundPages, 0, "Should not find any pages with no patterns")
		assert.Len(t, foundErrors, 0, "Should not have any errors with no patterns")
	})

	t.Run("Find single pattern", func(t *testing.T) {
		patterns := []string{"/pattern2"}
		var foundPages []*Page
		var foundErrors []error

		seq := store.FindByPatterns(ctx, ID("site1"), patterns...)
		for page, err := range seq {
			if err == nil {
				foundPages = append(foundPages, page)
			} else {
				foundErrors = append(foundErrors, err)
			}
		}

		assert.Len(t, foundPages, 1, "Should find 1 page")
		assert.Len(t, foundErrors, 0, "Should have 0 error values")
		assert.Equal(t, ID("page2"), foundPages[0].ID, "Should find correct page")
	})

	t.Run("Find with duplicate patterns", func(t *testing.T) {
		patterns := []string{"/pattern1", "/pattern1", "/pattern2"}
		var foundPages []*Page
		var foundErrors []error

		seq := store.FindByPatterns(ctx, ID("site1"), patterns...)
		for page, err := range seq {
			if err == nil {
				foundPages = append(foundPages, page)
			} else {
				foundErrors = append(foundErrors, err)
			}
		}

		assert.Len(t, foundPages, 2, "Should attempt to find 2 pages")
		assert.Len(t, foundErrors, 0, "Should have 0 error values")

		// Should find the same page multiple times for duplicates
		assert.Equal(t, ID("page1"), foundPages[0].ID, "Should find page1")
		assert.Equal(t, ID("page2"), foundPages[1].ID, "Should find page2")
	})

	t.Run("Find patterns with wrong site ID", func(t *testing.T) {
		patterns := []string{"/pattern1", "/pattern2"}
		var foundPages []*Page
		var foundErrors []error

		seq := store.FindByPatterns(ctx, ID("wrong-site"), patterns...)
		for page, err := range seq {
			if err == nil {
				foundPages = append(foundPages, page)
			} else {
				foundErrors = append(foundErrors, err)
			}
		}

		assert.Len(t, foundPages, 0, "Should attempt to find 0 pages")
		assert.Len(t, foundErrors, 0, "Should have 0 error values")
	})

	t.Run("Verify returned pages are copies", func(t *testing.T) {
		patterns := []string{"/pattern1"}
		var foundPages []*Page

		seq := store.FindByPatterns(ctx, ID("site1"), patterns...)
		for page, err := range seq {
			if err == nil {
				foundPages = append(foundPages, page)
			}
		}

		require.Len(t, foundPages, 1, "Should find 1 page")

		// Get original page for comparison
		originalPage, err := store.FindByID(ctx, foundPages[0].ID)
		require.NoError(t, err, "Should be able to find original page")

		// Modify the returned page
		foundPages[0].Title = "Modified Title"

		// Verify original page is unchanged
		originalPageAfter, err := store.FindByID(ctx, foundPages[0].ID)
		assert.NoError(t, err, "Should still be able to find original page")
		assert.NotEqual(t, "Modified Title", originalPageAfter.Title, "Original page should not be modified")
		assert.NotSame(t, originalPage, foundPages[0], "Should return a copy of the page")
	})
}

func TestMemoryPageStore_HelperFunctions(t *testing.T) {
	store := NewMemoryPageStore()
	ctx := context.Background()

	// Save test pages
	pages := []*Page{
		{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/test1", Alias: "alias1", Title: "Page 1"},
		{ID: ID("page2"), SiteID: ID("site1"), Pattern: "/test2", Alias: "alias2", Title: "Page 2"},
		{ID: ID("page3"), SiteID: ID("site1"), Pattern: "/test3", Alias: "alias3", Title: "Page 3"},
	}

	for _, page := range pages {
		err := store.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("deletePath helper function", func(t *testing.T) {
		// Get initial count
		initialCount := len(store.paths)
		assert.Greater(t, initialCount, 0, "Should have paths in store")

		// Delete middle page
		store.mu.Lock()
		store.deletePath(1) // Delete path for page2 (index 1)
		store.mu.Unlock()

		// Verify path was deleted
		assert.Len(t, store.paths, initialCount-1, "Path count should be reduced")
		assert.NotContains(t, store.paths, "site1-/test2", "Specific path should be deleted")

		// Verify other paths still exist
		assert.Contains(t, store.paths, "site1-/test1", "Other paths should still exist")
		assert.Contains(t, store.paths, "site1-/test3", "Other paths should still exist")
	})

	t.Run("deleteAlias helper function", func(t *testing.T) {
		// Get initial count
		initialCount := len(store.aliases)
		assert.Greater(t, initialCount, 0, "Should have aliases in store")

		// Delete middle page alias
		store.mu.Lock()
		store.deleteAlias(1) // Delete alias for page2 (index 1)
		store.mu.Unlock()

		// Verify alias was deleted
		assert.Len(t, store.aliases, initialCount-1, "Alias count should be reduced")
		assert.NotContains(t, store.aliases, "site1-alias2", "Specific alias should be deleted")

		// Verify other aliases still exist
		assert.Contains(t, store.aliases, "site1-alias1", "Other aliases should still exist")
		assert.Contains(t, store.aliases, "site1-alias3", "Other aliases should still exist")
	})

	t.Run("findByPath helper function", func(t *testing.T) {
		foundPage, err := store.findByPath(ID("site1"), "/test1")
		assert.NoError(t, err, "findByPath should find existing page")
		assert.Equal(t, ID("page1"), foundPage.ID, "Should find correct page")

		foundPage, err = store.findByPath(ID("site1"), "/non-existent")
		assert.Error(t, err, "findByPath should return error for non-existent page")
		assert.Nil(t, foundPage, "Should return nil for non-existent page")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})
}
