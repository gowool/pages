package pages

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromContext(t *testing.T) {
	t.Run("Context with Context", func(t *testing.T) {
		parent := context.Background()
		ctx, _ := NewContext(parent)

		c := FromContext(ctx)

		assert.NotNil(t, c, "FromContext should return Context")
		assert.Same(t, c, FromContext(ctx), "FromContext should return same instance")
	})

	t.Run("Context without Context", func(t *testing.T) {
		ctx := context.Background()

		c := FromContext(ctx)

		assert.Nil(t, c, "FromContext should return nil when Context not set")
	})

}

func TestNewContext(t *testing.T) {
	t.Run("Create new context", func(t *testing.T) {
		parent := context.Background()
		ctx, cancel := NewContext(parent)

		c := FromContext(ctx)

		require.NotNil(t, c, "NewContext should create Context")
		assert.NotNil(t, c.SEO(), "SEO should be initialized")
		assert.Equal(t, http.StatusOK, c.Status(), "Status should default to OK")
		assert.True(t, c.Guest(), "Guest should default to true")
		assert.False(t, c.Debug(), "Debug should default to false")
		assert.False(t, c.HasSite(), "HasSite should return false")
		assert.False(t, c.HasPage(), "HasPage should return false")
		assert.False(t, c.HasError(), "HasError should return false")

		cancel()

		c2 := FromContext(ctx)
		assert.NotNil(t, c2, "Context should still exist after cancel")
		assert.Same(t, c, c2, "Should be same instance")
	})

	t.Run("Cancel cleans up context", func(t *testing.T) {
		parent := context.Background()
		ctx, cancel := NewContext(parent)

		c := FromContext(ctx)
		c.SetDebug(true)
		c.SetGuest(false)
		c.SetStatus(http.StatusNotFound)

		cancel()

		assert.False(t, c.Debug(), "Debug should be reset")
		assert.True(t, c.Guest(), "Guest should be reset")
		assert.Equal(t, http.StatusOK, c.Status(), "Status should be reset")
	})

	t.Run("Pool reuse", func(t *testing.T) {
		ctx1, cancel1 := NewContext(context.Background())
		c1 := FromContext(ctx1)
		c1.SetDebug(true)
		cancel1()

		ctx2, cancel2 := NewContext(context.Background())
		c2 := FromContext(ctx2)

		assert.False(t, c2.Debug(), "Pooled context should be reset")

		cancel2()
	})
}

func TestContext_Reset(t *testing.T) {
	t.Run("Reset context with values", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetDebug(true)
		c.SetGuest(false)
		c.SetStatus(http.StatusNotFound)
		c.SetContent("<div>test</div>")
		c.SetSite(NewSite())
		c.SetPage(NewPage())
		c.SetError(errors.New("test error"))
		c.SEO().SetTitle("Test Title")

		c.Reset()

		assert.False(t, c.Debug(), "Debug should be false after reset")
		assert.True(t, c.Guest(), "Guest should be true after reset")
		assert.Equal(t, http.StatusOK, c.Status(), "Status should be OK after reset")
		assert.Equal(t, template.HTML(""), c.Content(), "Content should be empty after reset")
		assert.False(t, c.HasSite(), "Site should be nil after reset")
		assert.False(t, c.HasPage(), "Page should be nil after reset")
		assert.False(t, c.HasError(), "Error should be nil after reset")
		assert.Equal(t, "", c.SEO().Title(), "SEO should be reset")
	})

	t.Run("Reset new context", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.Reset()

		assert.Equal(t, http.StatusOK, c.Status(), "Status should be OK")
		assert.True(t, c.Guest(), "Guest should be true")
		assert.False(t, c.Debug(), "Debug should be false")
	})
}

func TestContext_SEO(t *testing.T) {
	t.Run("Lazy initialization", func(t *testing.T) {
		c := new(Context)

		assert.Nil(t, c.seo, "SEO should be nil initially")

		seo := c.SEO()

		assert.NotNil(t, seo, "SEO should be created on first access")
		assert.Same(t, seo, c.SEO(), "Should return same instance")
	})

	t.Run("Returns existing SEO", func(t *testing.T) {
		c := new(Context)
		c.seo = NewSEO()
		c.SEO().SetTitle("Test")

		seo := c.SEO()

		assert.Equal(t, "Test", seo.Title())
	})
}

func TestContext_Site(t *testing.T) {
	t.Run("Get site when set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		site := NewSite()
		c.SetSite(site)

		assert.Equal(t, site, c.Site())
	})

	t.Run("Get site when not set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		assert.Nil(t, c.Site())
	})
}

func TestContext_SetSite(t *testing.T) {
	t.Run("Set site with value", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		site := NewSite()
		site.Title = "Test Site"

		c.SetSite(site)

		assert.Equal(t, site, c.Site())
		assert.True(t, c.HasSite())
		assert.Equal(t, "Test Site", c.SEO().FirstTitle())
	})

	t.Run("Set site to nil", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetSite(NewSite())
		c.SetSite(nil)

		assert.Nil(t, c.Site())
		assert.False(t, c.HasSite())
	})
}

func TestContext_HasSite(t *testing.T) {
	t.Run("Has site when set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetSite(NewSite())

		assert.True(t, c.HasSite())
	})

	t.Run("No site when not set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		assert.False(t, c.HasSite())
	})

	t.Run("No site after nil set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetSite(NewSite())
		c.SetSite(nil)

		assert.False(t, c.HasSite())
	})
}

func TestContext_Page(t *testing.T) {
	t.Run("Get page when set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		page := NewPage()
		c.SetPage(page)

		assert.Equal(t, page, c.Page())
	})

	t.Run("Get page when not set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		assert.Nil(t, c.Page())
	})
}

func TestContext_SetPage(t *testing.T) {
	t.Run("Set page with site already set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		site := NewSite()
		site.ID = "site1"
		site.Title = "Test Site"
		c.SetSite(site)

		page := NewPage()
		page.Title = "Test Page"

		c.SetPage(page)

		assert.Equal(t, page, c.Page())
		assert.Equal(t, site, page.Site)
		assert.Equal(t, ID("site1"), page.SiteID)
		assert.Equal(t, "Test Site", c.SEO().FirstTitle())
	})

	t.Run("Set page without site set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		page := NewPage()
		page.Name = "Test Page"
		page.Site = NewSite()

		c.SetPage(page)

		assert.Equal(t, page, c.Page())
		assert.NotNil(t, page.Site)
	})

	t.Run("Set page with site and page.Site is nil", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		page := NewPage()
		page.Name = "Test Page"
		page.Site = nil

		c.SetPage(page)

		assert.Equal(t, site, page.Site)
		assert.Equal(t, ID("site1"), page.SiteID)
	})

	t.Run("Set page to nil", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetPage(NewPage())
		c.SetPage(nil)

		assert.Nil(t, c.Page())
		assert.False(t, c.HasPage())
	})

	t.Run("Set page with args", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		page := NewPage()
		page.Name = "Test Page"
		page.Pattern = "/test/{id}"

		c.SetPage(page, "{id}", "123")

		assert.Equal(t, page, c.Page())
	})
}

func TestContext_HasPage(t *testing.T) {
	t.Run("Has page when set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetPage(NewPage())

		assert.True(t, c.HasPage())
	})

	t.Run("No page when not set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		assert.False(t, c.HasPage())
	})

	t.Run("No page after nil set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetPage(NewPage())
		c.SetPage(nil)

		assert.False(t, c.HasPage())
	})
}

func TestContext_Error(t *testing.T) {
	t.Run("Get error when set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		err := errors.New("test error")
		c.SetError(err)

		assert.Equal(t, err, c.Error())
	})

	t.Run("Get error when not set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		assert.Nil(t, c.Error())
	})
}

func TestContext_SetError(t *testing.T) {
	t.Run("Set error", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		err := errors.New("test error")
		c.SetError(err)

		assert.Equal(t, err, c.Error())
		assert.True(t, c.HasError())
	})

	t.Run("Set error to nil", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetError(errors.New("test error"))
		c.SetError(nil)

		assert.Nil(t, c.Error())
		assert.False(t, c.HasError())
	})
}

func TestContext_HasError(t *testing.T) {
	t.Run("Has error when set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetError(errors.New("test"))

		assert.True(t, c.HasError())
	})

	t.Run("No error when not set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		assert.False(t, c.HasError())
	})

	t.Run("No error after nil set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetError(errors.New("test"))
		c.SetError(nil)

		assert.False(t, c.HasError())
	})
}

func TestContext_Debug(t *testing.T) {
	t.Run("Debug default", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		assert.False(t, c.Debug())
	})

	t.Run("Debug set to true", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetDebug(true)

		assert.True(t, c.Debug())
	})

	t.Run("Debug set to false", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetDebug(true)
		c.SetDebug(false)

		assert.False(t, c.Debug())
	})
}

func TestContext_SetDebug(t *testing.T) {
	t.Run("Set debug to true", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetDebug(true)

		assert.True(t, c.Debug())
	})

	t.Run("Set debug to false", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetDebug(false)

		assert.False(t, c.Debug())
	})
}

func TestContext_Guest(t *testing.T) {
	t.Run("Guest default", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		assert.True(t, c.Guest())
	})

	t.Run("Guest set to false", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetGuest(false)

		assert.False(t, c.Guest())
	})

	t.Run("Guest set to true", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetGuest(true)

		assert.True(t, c.Guest())
	})
}

func TestContext_SetGuest(t *testing.T) {
	t.Run("Set guest to true", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetGuest(true)

		assert.True(t, c.Guest())
	})

	t.Run("Set guest to false", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetGuest(false)

		assert.False(t, c.Guest())
	})
}

func TestContext_Status(t *testing.T) {
	t.Run("Status default", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		assert.Equal(t, http.StatusOK, c.Status())
	})

	t.Run("Status when set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetStatus(http.StatusNotFound)

		assert.Equal(t, http.StatusNotFound, c.Status())
	})

	t.Run("Status zero returns OK", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetStatus(0)

		assert.Equal(t, http.StatusOK, c.Status())
	})
}

func TestContext_SetStatus(t *testing.T) {
	t.Run("Set status", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetStatus(http.StatusNotFound)

		assert.Equal(t, http.StatusNotFound, c.Status())
	})

	t.Run("Set status to zero", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetStatus(0)

		assert.Equal(t, http.StatusOK, c.Status())
	})
}

func TestContext_Content(t *testing.T) {
	t.Run("Content default", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		assert.Equal(t, template.HTML(""), c.Content())
	})

	t.Run("Content when set", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		content := template.HTML("<div>test</div>")
		c.SetContent(content)

		assert.Equal(t, content, c.Content())
	})
}

func TestContext_SetContent(t *testing.T) {
	t.Run("Set content", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		content := template.HTML("<div>test</div>")
		c.SetContent(content)

		assert.Equal(t, content, c.Content())
	})

	t.Run("Set content to empty", func(t *testing.T) {
		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)

		c.SetContent("<div>test</div>")
		c.SetContent("")

		assert.Equal(t, template.HTML(""), c.Content())
	})
}

func TestContext_PoolConcurrency(t *testing.T) {
	t.Run("Concurrent context creation", func(t *testing.T) {
		done := make(chan bool, 100)

		for i := 0; i < 100; i++ {
			go func() {
				ctx, cancel := NewContext(context.Background())
				c := FromContext(ctx)
				c.SetDebug(true)
				cancel()
				done <- true
			}()
		}

		for i := 0; i < 100; i++ {
			<-done
		}

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		assert.False(t, c.Debug(), "Pooled context should be reset")
		cancel()
	})
}
