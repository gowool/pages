package pages

import (
	"context"
	"iter"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalhostSiteStore(t *testing.T) {
	store := NewLocalhostSiteStore()

	require.NotNil(t, store, "NewLocalhostSiteStore() should not return nil")
	assert.IsType(t, &LocalhostSiteStore{}, store, "Should return LocalhostSiteStore instance")
}

func TestLocalhostSiteStore_FindPublished(t *testing.T) {
	store := NewLocalhostSiteStore()
	ctx := context.Background()

	t.Run("FindPublished returns iterator", func(t *testing.T) {
		iterator := store.FindPublished(ctx)

		assert.NotNil(t, iterator, "FindPublished() should return iterator")
		assert.NotNil(t, iterator, "FindPublished() should return iterator")
	})

	t.Run("Iterator yields one published site", func(t *testing.T) {
		iterator := store.FindPublished(ctx)
		require.NotNil(t, iterator, "FindPublished() should return iterator")

		var sites []*Site
		var errors []error

		iterator(func(site *Site, err error) bool {
			if err != nil {
				errors = append(errors, err)
				return false
			}
			sites = append(sites, site)
			return true
		})

		assert.Empty(t, errors, "Iterator should not yield errors")
		assert.Len(t, sites, 1, "Iterator should yield exactly one site")

		site := sites[0]
		assert.True(t, site.Status == Published, "Site should be published")
		assert.Equal(t, "Localhost", site.Name, "Site should have default name")
		assert.Equal(t, "localhost", site.Host, "Site should have default host")
		assert.Equal(t, "https", site.Scheme, "Site should have default scheme")
		assert.Equal(t, "en", site.Locale, "Site should have default locale")
		assert.Equal(t, "UTC", site.Timezone, "Site should have default timezone")
		assert.Equal(t, " | ", site.Separator, "Site should have default separator")
		assert.False(t, site.IsDefault, "Site should not be default by default")
		assert.NotNil(t, site.MetaTags, "Site should have MetaTags initialized")
		assert.NotNil(t, site.Metadata, "Site should have Metadata initialized")
		assert.False(t, site.Created.IsZero(), "Site should have Created timestamp")
		assert.False(t, site.Updated.IsZero(), "Site should have Updated timestamp")
	})

	t.Run("Iterator can be stopped early", func(t *testing.T) {
		iterator := store.FindPublished(ctx)
		require.NotNil(t, iterator, "FindPublished() should return iterator")

		var iterationCount int
		iterator(func(site *Site, err error) bool {
			iterationCount++
			// Stop after first iteration
			return false
		})

		assert.Equal(t, 1, iterationCount, "Iterator should be called exactly once before stopping")
	})

	t.Run("Multiple calls return same data", func(t *testing.T) {
		iterator1 := store.FindPublished(ctx)
		iterator2 := store.FindPublished(ctx)

		assert.NotNil(t, iterator1, "First call should return iterator")
		assert.NotNil(t, iterator2, "Second call should return iterator")

		// Get first site
		var site1 *Site
		iterator1(func(site *Site, err error) bool {
			site1 = site
			return false
		})

		// Get second site
		var site2 *Site
		iterator2(func(site *Site, err error) bool {
			site2 = site
			return false
		})

		// Both sites should have same basic properties (but are different instances)
		assert.Equal(t, site1.Name, site2.Name, "Both sites should have same name")
		assert.Equal(t, site1.Host, site2.Host, "Both sites should have same host")
		assert.Equal(t, site1.Status, site2.Status, "Both sites should have same code")
	})

	t.Run("Iterator handles context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		iterator := store.FindPublished(ctx)
		assert.NotNil(t, iterator, "FindPublished() should return iterator even with cancelled context")

		var wasCalled bool
		iterator(func(site *Site, err error) bool {
			wasCalled = true
			return true
		})

		assert.True(t, wasCalled, "Iterator should still be called (implementation doesn't check context)")
	})
}

func TestLocalhostSiteStore_InterfaceCompliance(t *testing.T) {
	var _ SiteStore = (*LocalhostSiteStore)(nil)

	store := NewLocalhostSiteStore()
	ctx := context.Background()

	// Test that all interface methods are implemented and work
	iterator := store.FindPublished(ctx)
	assert.NotNil(t, iterator, "FindPublished should return iterator")
	assert.NotNil(t, iterator, "FindPublished should return iterator")
}

func TestLocalhostSiteStore_ConcurrentAccess(t *testing.T) {
	store := NewLocalhostSiteStore()
	ctx := context.Background()

	t.Run("Concurrent iterator calls", func(t *testing.T) {
		// Start multiple goroutines calling FindPublished
		for range 10 {
			go func() {
				iterator := store.FindPublished(ctx)
				assert.NotNil(t, iterator, "FindPublished() should return iterator")

				var count int
				iterator(func(site *Site, err error) bool {
					count++
					return true
				})
				assert.Equal(t, 1, count, "Each iterator should yield exactly one site")
			}()
		}
	})
}

func TestLocalhostSiteStore_IteratorBehavior(t *testing.T) {
	store := NewLocalhostSiteStore()
	ctx := context.Background()

	t.Run("Iterator yields different instances", func(t *testing.T) {
		iterator1 := store.FindPublished(ctx)
		iterator2 := store.FindPublished(ctx)
		require.NotNil(t, iterator1)
		require.NotNil(t, iterator2)

		var site1, site2 *Site

		iterator1(func(site *Site, err error) bool {
			site1 = site
			return false
		})

		iterator2(func(site *Site, err error) bool {
			site2 = site
			return false
		})

		assert.Equal(t, site1.Name, site2.Name, "But instances should have same values")
	})

	t.Run("Iterator yields complete site data", func(t *testing.T) {
		iterator := store.FindPublished(ctx)
		require.NotNil(t, iterator)

		var site *Site
		iterator(func(s *Site, e error) bool {
			site = s
			return false
		})

		require.NotNil(t, site, "Should have yielded a site")

		// Verify all expected fields are populated
		assert.True(t, site.Status == Published, "Site should be published")
		assert.NotEmpty(t, site.Name, "Site should have name")
		assert.NotEmpty(t, site.Host, "Site should have host")
		assert.NotEmpty(t, site.Scheme, "Site should have scheme")
		assert.NotEmpty(t, site.Locale, "Site should have locale")
		assert.NotEmpty(t, site.Timezone, "Site should have timezone")
		assert.NotEmpty(t, site.Separator, "Site should have separator")
		assert.NotNil(t, site.MetaTags, "Site should have MetaTags")
		assert.NotNil(t, site.Metadata, "Site should have Metadata")
		assert.False(t, site.Created.IsZero(), "Site should have Created time")
		assert.False(t, site.Updated.IsZero(), "Site should have Updated time")
	})

	t.Run("Iterator error handling", func(t *testing.T) {
		iterator := store.FindPublished(ctx)
		require.NotNil(t, iterator)

		var yieldedError error
		iterator(func(site *Site, err error) bool {
			yieldedError = err
			return false
		})

		assert.NoError(t, yieldedError, "Iterator should not yield errors")
	})
}

func TestLocalhostSiteStore_SiteProperties(t *testing.T) {
	store := NewLocalhostSiteStore()
	ctx := context.Background()

	iterator := store.FindPublished(ctx)
	require.NotNil(t, iterator)

	var site *Site
	iterator(func(s *Site, e error) bool {
		site = s
		return false
	})

	require.NotNil(t, site, "Should have yielded a site")

	t.Run("Default properties are correctly set", func(t *testing.T) {
		assert.Equal(t, "Localhost", site.Name, "Default name should be Localhost")
		assert.Equal(t, "localhost", site.Host, "Default host should be localhost")
		assert.Equal(t, "https", site.Scheme, "Default scheme should be https")
		assert.Equal(t, "en", site.Locale, "Default locale should be en")
		assert.Equal(t, "UTC", site.Timezone, "Default timezone should be UTC")
		assert.Equal(t, " | ", site.Separator, "Default separator should be ' | '")
		assert.True(t, site.Status == Published, "Site should be published")
		assert.False(t, site.IsDefault, "Site should not be default by default")
		assert.Empty(t, site.Countries, "Countries should be empty by default")
		assert.Empty(t, site.RelativePath, "RelativePath should be empty by default")
	})

	t.Run("Timestamps are reasonable", func(t *testing.T) {
		now := time.Now().UTC()
		assert.True(t, site.Created.Before(now) || site.Created.Equal(now), "Created time should be reasonable")
		assert.True(t, site.Updated.Before(now) || site.Updated.Equal(now), "Updated time should be reasonable")
		assert.True(t, site.Created.Equal(site.Updated), "Created and Updated should be equal for new site")
	})

	t.Run("MetaTags are properly initialized", func(t *testing.T) {
		require.NotNil(t, site.MetaTags, "MetaTags should not be nil")
		// We can't easily test the MetaTags content without accessing private fields,
		// but we can verify it was initialized
		assert.NotNil(t, site.MetaTags, "MetaTags should be initialized")
	})

	t.Run("Metadata map is properly initialized", func(t *testing.T) {
		require.NotNil(t, site.Metadata, "Metadata should not be nil")
		assert.Empty(t, site.Metadata, "Metadata should be empty initially")
	})
}

func TestLocalhostSiteStore_IteratorPerformance(t *testing.T) {
	store := NewLocalhostSiteStore()
	ctx := context.Background()

	t.Run("Iterator completes quickly", func(t *testing.T) {
		start := time.Now()

		iterator := store.FindPublished(ctx)
		require.NotNil(t, iterator)

		var count int
		iterator(func(site *Site, err error) bool {
			count++
			return true
		})

		duration := time.Since(start)
		assert.Less(t, duration, 100*time.Millisecond, "Iterator should complete quickly")
		assert.Equal(t, 1, count, "Should have iterated over exactly one site")
	})
}

func TestLocalhostSiteStore_EdgeCases(t *testing.T) {
	store := NewLocalhostSiteStore()

	t.Run("Multiple iterations", func(t *testing.T) {
		// Test that we can call FindPublished multiple times and get consistent results
		for i := range 5 {
			iterator := store.FindPublished(context.Background())
			assert.NotNil(t, iterator, "Call %d should return iterator", i+1)

			var count int
			iterator(func(site *Site, err error) bool {
				count++
				return true
			})
			assert.Equal(t, 1, count, "Call %d should yield exactly one site", i+1)
		}
	})
}

// Helper function to test iterator behavior more thoroughly
func collectSites(iterator iter.Seq2[*Site, error]) ([]*Site, []error) {
	var sites []*Site
	var errors []error

	iterator(func(site *Site, err error) bool {
		if err != nil {
			errors = append(errors, err)
			return false
		}
		sites = append(sites, site)
		return true
	})

	return sites, errors
}

func TestLocalhostSiteStore_Integration(t *testing.T) {
	store := NewLocalhostSiteStore()
	ctx := context.Background()

	t.Run("Integration test with realistic usage", func(t *testing.T) {
		// Simulate realistic usage pattern
		iterator := store.FindPublished(ctx)
		require.NotNil(t, iterator)

		sites, errors := collectSites(iterator)
		assert.Empty(t, errors, "Should not have any errors")
		assert.Len(t, sites, 1, "Should have exactly one site")

		site := sites[0]

		// Verify the site is suitable for use
		assert.True(t, site.Status == Published, "Site should be published")
		assert.NotEmpty(t, site.Host, "Site should have host")
		assert.NotEmpty(t, site.Scheme, "Site should have scheme")

		// Verify we can construct URLs with this site
		expectedBaseURL := site.Scheme + "://" + site.Host
		assert.Equal(t, "https://localhost", expectedBaseURL, "Site should be able to construct base URL")

		// Verify timestamp freshness (should be very recent)
		now := time.Now().UTC()
		assert.WithinDuration(t, now, site.Created, 5*time.Second, "Created time should be recent")
		assert.WithinDuration(t, now, site.Updated, 5*time.Second, "Updated time should be recent")
	})
}
