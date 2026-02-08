package pages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	t.Run("ErrSiteNotFound is defined", func(t *testing.T) {
		assert.NotNil(t, ErrSiteNotFound)
		assert.Equal(t, "site not found", ErrSiteNotFound.Error())
	})

	t.Run("ErrPageNotFound is defined", func(t *testing.T) {
		assert.NotNil(t, ErrPageNotFound)
		assert.Equal(t, "page not found", ErrPageNotFound.Error())
	})

	t.Run("ErrPageForbidden is defined", func(t *testing.T) {
		assert.NotNil(t, ErrPageForbidden)
		assert.Equal(t, "page forbidden", ErrPageForbidden.Error())
	})

	t.Run("ErrPageUnauthorized is defined", func(t *testing.T) {
		assert.NotNil(t, ErrPageUnauthorized)
		assert.Equal(t, "page unauthorized", ErrPageUnauthorized.Error())
	})

	t.Run("ErrUniqueViolation is defined", func(t *testing.T) {
		assert.NotNil(t, ErrUniqueViolation)
		assert.Equal(t, "unique violation", ErrUniqueViolation.Error())
	})

	t.Run("ErrTemplateEmpty is defined", func(t *testing.T) {
		assert.NotNil(t, ErrTemplateEmpty)
		assert.Equal(t, "template is empty", ErrTemplateEmpty.Error())
	})
}

func TestErrorVariableUniqueness(t *testing.T) {
	t.Run("all error variables are unique", func(t *testing.T) {
		errors := []error{
			ErrSiteNotFound,
			ErrPageNotFound,
			ErrPageForbidden,
			ErrPageUnauthorized,
			ErrUniqueViolation,
			ErrTemplateEmpty,
		}

		for i, e1 := range errors {
			for j, e2 := range errors {
				if i == j {
					continue
				}
				assert.NotEqual(t, e1, e2, "errors at indices %d and %d should be different", i, j)
			}
		}
	})
}
