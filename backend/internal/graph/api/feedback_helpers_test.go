package api

import (
	"context"
	"errors"
	"testing"

	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/stretchr/testify/assert"
)

func TestResolveFeedbackProjectID(t *testing.T) {
	submitted := "PR01ARZ3NDEKTSV4RRFFQ69G5FA"
	current := "PR01ARZ3NDEKTSV4RRFFQ69G5FB"

	t.Run("keeps the project the client sent", func(t *testing.T) {
		projects := mocks.NewMockProjectIDProvider(t)

		got := resolveFeedbackProjectID(context.Background(), projects, &submitted)

		assert.Equal(t, submitted, got)
	})

	t.Run("falls back to the active project when none is sent", func(t *testing.T) {
		projects := mocks.NewMockProjectIDProvider(t)
		projects.EXPECT().GetCurrentProjectID(context.Background()).Return(current, nil)

		got := resolveFeedbackProjectID(context.Background(), projects, nil)

		assert.Equal(t, current, got)
	})

	t.Run("falls back when the client sends an empty project", func(t *testing.T) {
		projects := mocks.NewMockProjectIDProvider(t)
		projects.EXPECT().GetCurrentProjectID(context.Background()).Return(current, nil)
		empty := ""

		got := resolveFeedbackProjectID(context.Background(), projects, &empty)

		assert.Equal(t, current, got)
	})

	t.Run("returns empty when the active project cannot be resolved", func(t *testing.T) {
		projects := mocks.NewMockProjectIDProvider(t)
		projects.EXPECT().GetCurrentProjectID(context.Background()).Return("", errors.New("no setting"))

		got := resolveFeedbackProjectID(context.Background(), projects, nil)

		assert.Empty(t, got)
	})

	t.Run("returns empty when no provider is configured", func(t *testing.T) {
		got := resolveFeedbackProjectID(context.Background(), nil, nil)

		assert.Empty(t, got)
	})
}
