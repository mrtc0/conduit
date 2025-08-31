package testutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockClock_NewTicker(t *testing.T) {
	t.Parallel()

	duration := 30 * time.Second

	doneTest := make(chan struct{})
	defer func() { <-doneTest }() // block until the test is done

	mockClock := NewMockClock()
	ticker := mockClock.NewTicker(duration)
	assert.NotNil(t, ticker.C)

	quit := make(chan struct{})

	go func() {
		defer close(doneTest)

		for {
			select {
			case <-ticker.C:
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	current := mockClock.Now()
	mockClock.Add(30 * time.Second)
	assert.Equal(t, current.Add(30*time.Second), mockClock.Now())

	mockClock.Add(30 * time.Second)
	assert.Equal(t, current.Add(60*time.Second), mockClock.Now())

	close(quit)
}
