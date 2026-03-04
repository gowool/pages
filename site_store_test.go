package pages

import (
	"context"
	"testing"

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

	sites, err := store.FindPublished(ctx)

	assert.NoError(t, err)
	assert.Len(t, sites, 1)

	site := sites[0]
	assert.True(t, site.Status == Published, "Site should be published")
	assert.Equal(t, "Localhost", site.Name, "Site should have default name")
	assert.Equal(t, "localhost", site.Host, "Site should have default host")
	assert.Equal(t, "https", site.Scheme, "Site should have default scheme")
	assert.Equal(t, "en", site.Locale, "Site should have default locale")
	assert.Equal(t, "UTC", site.Timezone, "Site should have default timezone")
	assert.Equal(t, " | ", site.Separator, "Site should have default separator")
	assert.False(t, site.IsDefault, "Site should not be default by default")
	assert.NotNil(t, site.Meta, "Site should have Meta initialized")
	assert.False(t, site.Created.IsZero(), "Site should have Created timestamp")
	assert.False(t, site.Updated.IsZero(), "Site should have Updated timestamp")
}
