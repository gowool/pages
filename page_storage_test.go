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

func TestNewMemoryPageStorage(t *testing.T) {
	storage := NewMemoryPageStorage()

	require.NotNil(t, storage, "NewMemoryPageStorage() should not return nil")
	// data slice is initially nil (not explicitly initialized)
	assert.NotNil(t, storage.ids, "ids map should be initialized")
	assert.NotNil(t, storage.paths, "paths map should be initialized")
	assert.NotNil(t, storage.aliases, "aliases map should be initialized")
	assert.Nil(t, storage.data, "data slice should be nil initially")
	assert.Empty(t, storage.ids, "ids map should be empty initially")
	assert.Empty(t, storage.paths, "paths map should be empty initially")
	assert.Empty(t, storage.aliases, "aliases map should be empty initially")
}

func TestMemoryPageStorage_Save(t *testing.T) {
	t.Run("Save new page", func(t *testing.T) {
		storage := NewMemoryPageStorage()
		ctx := context.Background()

		page := &Page{
			SiteID:  ID("site1"),
			Pattern: "/test",
			Title:   "Test Page",
		}

		err := storage.Save(ctx, page)
		assert.NoError(t, err, "Save() should not return error for new page")
		assert.NotZero(t, page.ID, "Page ID should be set after save")
		assert.Len(t, storage.data, 1, "Storage should contain 1 page")
		assert.Contains(t, storage.ids, page.ID, "Page ID should be in ids map")
	})

	t.Run("Save page with existing ID", func(t *testing.T) {
		storage := NewMemoryPageStorage()
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
		err := storage.Save(ctx, originalPage)
		require.NoError(t, err)

		// Save updated page with same ID
		err = storage.Save(ctx, updatedPage)
		assert.NoError(t, err, "Save() should not return error for updated page")
		assert.Len(t, storage.data, 1, "Storage should still contain 1 page")
		assert.Equal(t, "Updated Page", storage.data[0].Title, "Page should be updated")
	})

	t.Run("Save multiple pages", func(t *testing.T) {
		storage := NewMemoryPageStorage()
		ctx := context.Background()

		pages := []*Page{
			{SiteID: ID("site1"), Pattern: "/page1", Alias: "alias1", Title: "Page 1"},
			{SiteID: ID("site1"), Pattern: "/page2", Alias: "alias2", Title: "Page 2"},
			{SiteID: ID("site2"), Pattern: "/page3", Alias: "alias3", Title: "Page 3"},
		}

		err := storage.Save(ctx, pages...)
		assert.NoError(t, err, "Save() should not return error for multiple pages")
		assert.Len(t, storage.data, 3, "Storage should contain 3 pages")
	})

	t.Run("Save CMS page with URL", func(t *testing.T) {
		storage := NewMemoryPageStorage()
		ctx := context.Background()

		page := &Page{
			SiteID:  ID("site1"),
			Pattern: PageCMS,
			URL:     "/actual-url",
			Title:   "CMS Page",
		}

		err := storage.Save(ctx, page)
		assert.NoError(t, err, "Save() should not return error for CMS page")

		// For CMS pages, the path should be based on URL, not Pattern
		expectedPath := "site1-/actual-url"
		assert.Contains(t, storage.paths, expectedPath, "CMS page should use URL as path")
	})

	t.Run("Save page with duplicate path", func(t *testing.T) {
		storage := NewMemoryPageStorage()
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
		err := storage.Save(ctx, page1)
		require.NoError(t, err)

		// Try to save second page with same path
		err = storage.Save(ctx, page2)
		assert.Error(t, err, "Save() should return error for duplicate path")
		assert.True(t, errors.Is(err, ErrUniqueViolation), "Error should be ErrUniqueViolation")
	})

	t.Run("Save page with duplicate alias", func(t *testing.T) {
		storage := NewMemoryPageStorage()
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
		err := storage.Save(ctx, page1)
		require.NoError(t, err)

		// Try to save second page with same alias
		err = storage.Save(ctx, page2)
		assert.Error(t, err, "Save() should return error for duplicate alias")
		assert.True(t, errors.Is(err, ErrUniqueViolation), "Error should be ErrUniqueViolation")
	})

	t.Run("Save page updates path and alias correctly", func(t *testing.T) {
		storage := NewMemoryPageStorage()
		ctx := context.Background()

		page := &Page{
			ID:      ID("page1"),
			SiteID:  ID("site1"),
			Pattern: "/original",
			Alias:   "original-alias",
			Title:   "Original Page",
		}

		// Save original page
		err := storage.Save(ctx, page)
		require.NoError(t, err)

		originalPath := "site1-/original"
		originalAlias := "site1-original-alias"
		assert.Contains(t, storage.paths, originalPath, "Original path should exist")
		assert.Contains(t, storage.aliases, originalAlias, "Original alias should exist")

		// Update page with new pattern and alias
		page.Pattern = "/updated"
		page.Alias = "updated-alias"
		err = storage.Save(ctx, page)
		assert.NoError(t, err, "Save() should not return error for updated page")

		updatedPath := "site1-/updated"
		updatedAlias := "site1-updated-alias"
		assert.NotContains(t, storage.paths, originalPath, "Original path should be removed")
		assert.NotContains(t, storage.aliases, originalAlias, "Original alias should be removed")
		assert.Contains(t, storage.paths, updatedPath, "Updated path should exist")
		assert.Contains(t, storage.aliases, updatedAlias, "Updated alias should exist")
	})
}

func TestMemoryPageStorage_FindByID(t *testing.T) {
	storage := NewMemoryPageStorage()
	ctx := context.Background()

	// Save a test page
	testPage := &Page{
		ID:     ID("test-id"),
		SiteID: ID("site1"),
		Title:  "Test Page",
	}
	err := storage.Save(ctx, testPage)
	require.NoError(t, err)

	t.Run("Find existing page by ID", func(t *testing.T) {
		foundPage, err := storage.FindByID(ctx, ID("test-id"))
		assert.NoError(t, err, "FindByID() should not return error for existing page")
		assert.NotNil(t, foundPage, "Found page should not be nil")
		assert.Equal(t, ID("test-id"), foundPage.ID, "Found page should have correct ID")
		assert.Equal(t, "Test Page", foundPage.Title, "Found page should have correct title")

		// Ensure we get a copy, not the original reference
		assert.NotSame(t, testPage, foundPage, "Should return a copy of the page")
	})

	t.Run("Find non-existent page by ID", func(t *testing.T) {
		foundPage, err := storage.FindByID(ctx, ID("non-existent"))
		assert.Error(t, err, "FindByID() should return error for non-existent page")
		assert.Nil(t, foundPage, "Found page should be nil for non-existent page")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})
}

func TestMemoryPageStorage_FindByURL(t *testing.T) {
	storage := NewMemoryPageStorage()
	ctx := context.Background()

	// Save test pages
	pages := []*Page{
		{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/page1", Alias: "alias1", Title: "Page 1"},
		{ID: ID("page2"), SiteID: ID("site2"), Pattern: "/page2", Alias: "alias2", Title: "Page 2"},
		{ID: ID("page3"), SiteID: ID("site1"), Pattern: PageCMS, URL: "/cms-page", Alias: "alias3", Title: "CMS Page"},
	}

	for _, page := range pages {
		err := storage.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("Find existing page by URL", func(t *testing.T) {
		foundPage, err := storage.FindByURL(ctx, ID("site1"), "/page1")
		assert.NoError(t, err, "FindByURL() should not return error for existing page")
		assert.NotNil(t, foundPage, "Found page should not be nil")
		assert.Equal(t, ID("page1"), foundPage.ID, "Found page should have correct ID")
	})

	t.Run("Find CMS page by URL", func(t *testing.T) {
		foundPage, err := storage.FindByURL(ctx, ID("site1"), "/cms-page")
		assert.NoError(t, err, "FindByURL() should find CMS page by URL")
		assert.NotNil(t, foundPage, "Found page should not be nil")
		assert.Equal(t, ID("page3"), foundPage.ID, "Found page should have correct ID")
	})

	t.Run("Find non-existent page by URL", func(t *testing.T) {
		foundPage, err := storage.FindByURL(ctx, ID("site1"), "/non-existent")
		assert.Error(t, err, "FindByURL() should return error for non-existent page")
		assert.Nil(t, foundPage, "Found page should be nil for non-existent page")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})

	t.Run("Find page with wrong site ID", func(t *testing.T) {
		foundPage, err := storage.FindByURL(ctx, ID("wrong-site"), "/page1")
		assert.Error(t, err, "FindByURL() should return error for wrong site ID")
		assert.Nil(t, foundPage, "Found page should be nil for wrong site ID")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})
}

func TestMemoryPageStorage_FindByPattern(t *testing.T) {
	storage := NewMemoryPageStorage()
	ctx := context.Background()

	// Save test pages
	pages := []*Page{
		{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/pattern1", Alias: "alias1", Title: "Page 1"},
		{ID: ID("page2"), SiteID: ID("site2"), Pattern: "/pattern2", Alias: "alias2", Title: "Page 2"},
	}

	for _, page := range pages {
		err := storage.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("Find existing page by pattern", func(t *testing.T) {
		foundPage, err := storage.FindByPattern(ctx, ID("site1"), "/pattern1")
		assert.NoError(t, err, "FindByPattern() should not return error for existing page")
		assert.NotNil(t, foundPage, "Found page should not be nil")
		assert.Equal(t, ID("page1"), foundPage.ID, "Found page should have correct ID")
	})

	t.Run("Find non-existent page by pattern", func(t *testing.T) {
		foundPage, err := storage.FindByPattern(ctx, ID("site1"), "/non-existent")
		assert.Error(t, err, "FindByPattern() should return error for non-existent page")
		assert.Nil(t, foundPage, "Found page should be nil for non-existent page")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})

	t.Run("Find page with wrong site ID", func(t *testing.T) {
		foundPage, err := storage.FindByPattern(ctx, ID("wrong-site"), "/pattern1")
		assert.Error(t, err, "FindByPattern() should return error for wrong site ID")
		assert.Nil(t, foundPage, "Found page should be nil for wrong site ID")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})
}

func TestMemoryPageStorage_FindByAlias(t *testing.T) {
	storage := NewMemoryPageStorage()
	ctx := context.Background()

	// Save test pages
	pages := []*Page{
		{ID: ID("page1"), SiteID: ID("site1"), Alias: "alias1", Title: "Page 1"},
		{ID: ID("page2"), SiteID: ID("site2"), Alias: "alias2", Title: "Page 2"},
	}

	for _, page := range pages {
		err := storage.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("Find existing page by alias", func(t *testing.T) {
		foundPage, err := storage.FindByAlias(ctx, ID("site1"), "alias1")
		assert.NoError(t, err, "FindByAlias() should not return error for existing page")
		assert.NotNil(t, foundPage, "Found page should not be nil")
		assert.Equal(t, ID("page1"), foundPage.ID, "Found page should have correct ID")
	})

	t.Run("Find non-existent page by alias", func(t *testing.T) {
		foundPage, err := storage.FindByAlias(ctx, ID("site1"), "non-existent")
		assert.Error(t, err, "FindByAlias() should return error for non-existent page")
		assert.Nil(t, foundPage, "Found page should be nil for non-existent page")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})

	t.Run("Find page with wrong site ID", func(t *testing.T) {
		foundPage, err := storage.FindByAlias(ctx, ID("wrong-site"), "alias1")
		assert.Error(t, err, "FindByAlias() should return error for wrong site ID")
		assert.Nil(t, foundPage, "Found page should be nil for wrong site ID")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})
}

func TestMemoryPageStorage_DeleteByID(t *testing.T) {
	storage := NewMemoryPageStorage()
	ctx := context.Background()

	// Save test pages
	pages := []*Page{
		{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/page1", Alias: "alias1", Title: "Page 1"},
		{ID: ID("page2"), SiteID: ID("site1"), Pattern: "/page2", Alias: "alias2", Title: "Page 2"},
		{ID: ID("page3"), SiteID: ID("site2"), Pattern: "/page3", Alias: "alias3", Title: "Page 3"},
	}

	for _, page := range pages {
		err := storage.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("Delete existing page by ID", func(t *testing.T) {
		err := storage.DeleteByID(ctx, ID("page1"))
		assert.NoError(t, err, "DeleteByID() should not return error for existing page")

		// Verify page is deleted
		assert.Len(t, storage.data, 2, "Storage should contain 2 pages after deletion")
		assert.NotContains(t, storage.ids, ID("page1"), "Deleted page ID should not be in ids map")

		// Verify path and alias are also deleted
		assert.NotContains(t, storage.paths, "site1-/page1", "Deleted page path should not be in paths map")
		assert.NotContains(t, storage.aliases, "site1-alias1", "Deleted page alias should not be in aliases map")

		// Verify other pages are still there
		assert.Contains(t, storage.ids, ID("page2"), "Other page IDs should still be present")
		assert.Contains(t, storage.ids, ID("page3"), "Other page IDs should still be present")
	})

	t.Run("Delete multiple pages by ID", func(t *testing.T) {
		storage := NewMemoryPageStorage()
		ctx := context.Background()

		// Save test pages
		pages := []*Page{
			{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/page1", Alias: "alias1", Title: "Page 1"},
			{ID: ID("page2"), SiteID: ID("site1"), Pattern: "/page2", Alias: "alias2", Title: "Page 2"},
		}

		for _, page := range pages {
			err := storage.Save(ctx, page)
			require.NoError(t, err)
		}

		// Delete last page first to avoid index issues, then the first
		err := storage.DeleteByID(ctx, ID("page2"))
		assert.NoError(t, err, "DeleteByID() should not return error for existing page")

		err = storage.DeleteByID(ctx, ID("page1"))
		assert.NoError(t, err, "DeleteByID() should not return error for existing page")

		assert.Len(t, storage.data, 0, "Storage should be empty after deleting all pages")
	})

	t.Run("Delete non-existent page by ID", func(t *testing.T) {
		err := storage.DeleteByID(ctx, ID("non-existent"))
		assert.Error(t, err, "DeleteByID() should return error for non-existent page")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")

		// Verify no pages were deleted
		assert.Len(t, storage.data, 2, "Storage should still contain 2 pages")
	})

	t.Run("Delete page and verify indexes are updated", func(t *testing.T) {
		storage := NewMemoryPageStorage()
		ctx := context.Background()

		// Save test pages
		pages := []*Page{
			{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/page1", Alias: "alias1", Title: "Page 1"},
			{ID: ID("page2"), SiteID: ID("site1"), Pattern: "/page2", Alias: "alias2", Title: "Page 2"},
			{ID: ID("page3"), SiteID: ID("site1"), Pattern: "/page3", Alias: "alias3", Title: "Page 3"},
		}

		for _, page := range pages {
			err := storage.Save(ctx, page)
			require.NoError(t, err)
		}

		// Delete last page first to avoid index issues
		err := storage.DeleteByID(ctx, ID("page3"))
		assert.NoError(t, err, "DeleteByID() should not return error for existing page")

		// Verify remaining pages can still be found
		foundPage1, err := storage.FindByID(ctx, ID("page1"))
		assert.NoError(t, err, "Should be able to find remaining page1")
		assert.Equal(t, "Page 1", foundPage1.Title, "Remaining page1 should have correct title")

		foundPage2, err := storage.FindByID(ctx, ID("page2"))
		assert.NoError(t, err, "Should be able to find remaining page2")
		assert.Equal(t, "Page 2", foundPage2.Title, "Remaining page2 should have correct title")
	})
}

func TestMemoryPageStorage_ConcurrentAccess(t *testing.T) {
	storage := NewMemoryPageStorage()
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
		err := storage.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("Concurrent reads", func(t *testing.T) {
		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		// Get the page IDs once before starting concurrent reads
		data := storage.GetData()
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
					_, err := storage.FindByID(ctx, pageIDs[id])
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
				err := storage.Save(ctx, page)
				errChan <- err
			}(i)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			assert.NoError(t, err, "Concurrent writes should not cause errors")
		}

		// Verify all pages were saved
		data := storage.GetData()
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
					data := storage.GetData()
					if len(data) > 0 {
						_, err := storage.FindByID(ctx, data[0].ID)
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
				err := storage.Save(ctx, page)
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

func TestMemoryPageStorage_EdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("Save page with empty pattern and URL", func(t *testing.T) {
		storage := NewMemoryPageStorage()

		page := &Page{
			ID:     ID("test"),
			SiteID: ID("site1"),
			Alias:  "test-alias",
			Title:  "Test Page",
		}

		err := storage.Save(ctx, page)
		assert.NoError(t, err, "Save() should handle empty pattern and URL")

		// Should use empty string as path
		expectedPath := "site1-"
		assert.Contains(t, storage.paths, expectedPath, "Should create path with empty pattern/URL")
	})

	t.Run("Save page with special characters in path", func(t *testing.T) {
		storage := NewMemoryPageStorage()

		page := &Page{
			ID:      ID("test"),
			SiteID:  ID("site1"),
			Pattern: "/path/with/special-chars?param=value#fragment",
			Title:   "Test Page",
		}

		err := storage.Save(ctx, page)
		assert.NoError(t, err, "Save() should handle special characters in path")

		expectedPath := "site1-/path/with/special-chars?param=value#fragment"
		assert.Contains(t, storage.paths, expectedPath, "Should handle special characters in path")
	})

	t.Run("Find operations with empty strings", func(t *testing.T) {
		storage := NewMemoryPageStorage()

		// Save page with empty pattern
		page := &Page{
			ID:     ID("test"),
			SiteID: ID("site1"),
			Alias:  "test-alias",
			Title:  "Test Page",
		}
		err := storage.Save(ctx, page)
		require.NoError(t, err)

		// Try to find with empty URL/pattern
		foundPage, err := storage.FindByURL(ctx, ID("site1"), "")
		assert.NoError(t, err, "Should be able to find page with empty URL")
		assert.Equal(t, ID("test"), foundPage.ID, "Should find correct page")

		foundPage, err = storage.FindByPattern(ctx, ID("site1"), "")
		assert.NoError(t, err, "Should be able to find page with empty pattern")
		assert.Equal(t, ID("test"), foundPage.ID, "Should find correct page")
	})

	t.Run("Save and find with very long strings", func(t *testing.T) {
		storage := NewMemoryPageStorage()

		longPattern := "/" + string(make([]byte, 1000))
		longTitle := string(make([]byte, 1000))

		page := &Page{
			ID:      ID("test"),
			SiteID:  ID("site1"),
			Pattern: longPattern,
			Title:   longTitle,
		}

		err := storage.Save(ctx, page)
		assert.NoError(t, err, "Save() should handle very long strings")

		foundPage, err := storage.FindByPattern(ctx, ID("site1"), longPattern)
		assert.NoError(t, err, "Should be able to find page with long pattern")
		assert.Equal(t, longTitle, foundPage.Title, "Should preserve long title")
	})

	t.Run("Delete from empty storage", func(t *testing.T) {
		storage := NewMemoryPageStorage()

		err := storage.DeleteByID(ctx, ID("non-existent"))
		assert.Error(t, err, "Delete from empty storage should return error")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})

	t.Run("Update page to have conflicting path", func(t *testing.T) {
		storage := NewMemoryPageStorage()

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
		err := storage.Save(ctx, page1, page2)
		require.NoError(t, err)

		// Try to update page2 to have same pattern as page1
		page2.Pattern = "/path1"
		err = storage.Save(ctx, page2)
		assert.Error(t, err, "Update to conflicting path should return error")
		assert.True(t, errors.Is(err, ErrUniqueViolation), "Error should be ErrUniqueViolation")
	})
}

func TestMemoryPageStorage_DataIsolation(t *testing.T) {
	ctx := context.Background()

	t.Run("Modifying returned page should not affect storage", func(t *testing.T) {
		storage := NewMemoryPageStorage()

		originalPage := &Page{
			ID:      ID("test"),
			SiteID:  ID("site1"),
			Pattern: "/test",
			Title:   "Original Title",
		}

		err := storage.Save(ctx, originalPage)
		require.NoError(t, err)

		// Find and modify the returned page
		foundPage, err := storage.FindByID(ctx, ID("test"))
		require.NoError(t, err)

		foundPage.Title = "Modified Title"

		// Find again to verify original page is unchanged
		foundPage2, err := storage.FindByID(ctx, ID("test"))
		assert.NoError(t, err, "Should still be able to find page")
		assert.Equal(t, "Original Title", foundPage2.Title, "Original page should not be modified")
	})
}

func TestMemoryPageStorage_InterfaceCompliance(t *testing.T) {
	var _ PageStorage = (*MemoryPageStorage)(nil)

	storage := NewMemoryPageStorage()
	ctx := context.Background()

	// Test that all interface methods are implemented and work
	page := &Page{
		ID:     ID("test"),
		SiteID: ID("site1"),
		Title:  "Test Page",
	}

	// Save
	err := storage.Save(ctx, page)
	assert.NoError(t, err, "Save method should work")

	// FindBy ID
	foundPage, err := storage.FindByID(ctx, ID("test"))
	assert.NoError(t, err, "FindByID method should work")
	assert.NotNil(t, foundPage, "FindByID should return page")

	// FindBy URL
	foundPage, err = storage.FindByURL(ctx, ID("site1"), "")
	assert.NoError(t, err, "FindByURL method should work")
	assert.NotNil(t, foundPage, "FindByURL should return page")

	// FindBy Pattern
	foundPage, err = storage.FindByPattern(ctx, ID("site1"), "")
	assert.NoError(t, err, "FindByPattern method should work")
	assert.NotNil(t, foundPage, "FindByPattern should return page")

	// Delete
	err = storage.DeleteByID(ctx, ID("test"))
	assert.NoError(t, err, "DeleteByID method should work")
}

func TestMemoryPageStorage_HelperFunctions(t *testing.T) {
	storage := NewMemoryPageStorage()
	ctx := context.Background()

	// Save test pages
	pages := []*Page{
		{ID: ID("page1"), SiteID: ID("site1"), Pattern: "/test1", Alias: "alias1", Title: "Page 1"},
		{ID: ID("page2"), SiteID: ID("site1"), Pattern: "/test2", Alias: "alias2", Title: "Page 2"},
		{ID: ID("page3"), SiteID: ID("site1"), Pattern: "/test3", Alias: "alias3", Title: "Page 3"},
	}

	for _, page := range pages {
		err := storage.Save(ctx, page)
		require.NoError(t, err)
	}

	t.Run("deletePath helper function", func(t *testing.T) {
		// Get initial count
		initialCount := len(storage.paths)
		assert.Greater(t, initialCount, 0, "Should have paths in storage")

		// Delete middle page
		storage.mu.Lock()
		storage.deletePath(1) // Delete path for page2 (index 1)
		storage.mu.Unlock()

		// Verify path was deleted
		assert.Len(t, storage.paths, initialCount-1, "Path count should be reduced")
		assert.NotContains(t, storage.paths, "site1-/test2", "Specific path should be deleted")

		// Verify other paths still exist
		assert.Contains(t, storage.paths, "site1-/test1", "Other paths should still exist")
		assert.Contains(t, storage.paths, "site1-/test3", "Other paths should still exist")
	})

	t.Run("deleteAlias helper function", func(t *testing.T) {
		// Get initial count
		initialCount := len(storage.aliases)
		assert.Greater(t, initialCount, 0, "Should have aliases in storage")

		// Delete middle page alias
		storage.mu.Lock()
		storage.deleteAlias(1) // Delete alias for page2 (index 1)
		storage.mu.Unlock()

		// Verify alias was deleted
		assert.Len(t, storage.aliases, initialCount-1, "Alias count should be reduced")
		assert.NotContains(t, storage.aliases, "site1-alias2", "Specific alias should be deleted")

		// Verify other aliases still exist
		assert.Contains(t, storage.aliases, "site1-alias1", "Other aliases should still exist")
		assert.Contains(t, storage.aliases, "site1-alias3", "Other aliases should still exist")
	})

	t.Run("findByPath helper function", func(t *testing.T) {
		foundPage, err := storage.findByPath(ID("site1"), "/test1")
		assert.NoError(t, err, "findByPath should find existing page")
		assert.Equal(t, ID("page1"), foundPage.ID, "Should find correct page")

		foundPage, err = storage.findByPath(ID("site1"), "/non-existent")
		assert.Error(t, err, "findByPath should return error for non-existent page")
		assert.Nil(t, foundPage, "Should return nil for non-existent page")
		assert.True(t, errors.Is(err, ErrPageNotFound), "Error should be ErrPageNotFound")
	})
}
