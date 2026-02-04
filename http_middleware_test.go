package pages

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelayedWriter(t *testing.T) {
	t.Run("Reset initializes state", func(t *testing.T) {
		w := &delayedWriter{}
		rw := httptest.NewRecorder()

		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("test"))

		w.reset(rw)

		assert.Equal(t, rw, w.ResponseWriter)
		assert.Equal(t, 0, w.buffer.Len())
		assert.False(t, w.commited)
		assert.Equal(t, http.StatusOK, w.status)
	})

	t.Run("WriteHeader tracks status", func(t *testing.T) {
		w := &delayedWriter{}

		w.WriteHeader(http.StatusCreated)

		assert.Equal(t, http.StatusCreated, w.status)
		assert.True(t, w.commited)
	})

	t.Run("Write buffers data", func(t *testing.T) {
		w := &delayedWriter{}
		data := []byte("test data")

		n, err := w.Write(data)

		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, "test data", w.buffer.String())
		assert.True(t, w.commited)
	})

	t.Run("Write commits if not already committed", func(t *testing.T) {
		w := &delayedWriter{}

		n, err := w.Write([]byte("test"))

		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, http.StatusOK, w.status)
		assert.True(t, w.commited)
	})

	t.Run("Unwrap returns ResponseWriter", func(t *testing.T) {
		rw := httptest.NewRecorder()
		w := &delayedWriter{ResponseWriter: rw}

		unwrapped := w.Unwrap()

		assert.Equal(t, rw, unwrapped)
	})
}
