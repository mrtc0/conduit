package testutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	t.Parallel()

	testCases := []int{10, 20, 50, 100}
	for _, length := range testCases {
		length := length

		t.Run(fmt.Sprintf("RandomString length %d", length), func(t *testing.T) {
			t.Parallel()

			result := RandomString(length)
			assert.Equal(t, length, len(result))
		})
	}
}
