package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentTime(t *testing.T) {
	r := &queryResolver{&Resolver{}}

	before := time.Now()
	got, err := r.CurrentTime(context.Background())
	after := time.Now()

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.False(t, got.Before(before), "currentTime should not be before the call")
	assert.False(t, got.After(after), "currentTime should not be after the call")
	assert.Equal(t, time.UTC, got.Location())
}
