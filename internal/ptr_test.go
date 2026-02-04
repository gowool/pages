package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRef(t *testing.T) {
	t.Run("Ref creates pointer to value", func(t *testing.T) {
		strValue := "test"
		strPtr := Ref(strValue)
		assert.NotNil(t, strPtr)
		assert.Equal(t, "test", *strPtr)

		intValue := 42
		intPtr := Ref(intValue)
		assert.NotNil(t, intPtr)
		assert.Equal(t, 42, *intPtr)

		boolValue := true
		boolPtr := Ref(boolValue)
		assert.NotNil(t, boolPtr)
		assert.True(t, *boolPtr)
	})
}
