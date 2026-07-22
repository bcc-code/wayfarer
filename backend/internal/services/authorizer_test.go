package services

import (
	"context"
	"testing"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCan_DispatchParity verifies that RoleService.Can dispatches each Action to
// the corresponding underlying method and returns the identical result.
func TestCan_DispatchParity(t *testing.T) {
	ctx := context.Background()
	actorID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// mockAdmin wires the querier so the actor is a superadmin, which every
	// Manage*/AssignRole check short-circuits on.
	mockAdmin := func(m *mocks.MockRoleQuerier) {
		m.On("HasRole", ctx, sqlc.HasRoleParams{UserID: actorID, Role: string(RoleSuperAdmin)}).Return(true, nil)
		m.On("HasRole", ctx, sqlc.HasRoleParams{UserID: actorID, Role: string(RoleAdmin)}).Return(true, nil).Maybe()
	}

	t.Run("access_user matches CanAccessUser (self)", func(t *testing.T) {
		mockQueries := mocks.NewMockRoleQuerier(t)
		service := NewRoleService(mockQueries, newTestCache())
		res := Resource{TargetUserID: actorID} // self-access short-circuits
		got, err := service.Can(ctx, actorID, ActionAccessUser, res)
		require.NoError(t, err)
		assert.Equal(t, service.CanAccessUser(ctx, actorID, res.TargetUserID, res.TargetUserChurchID), got)
		assert.True(t, got)
	})

	t.Run("manage_project matches CanManageProject", func(t *testing.T) {
		mockQueries := mocks.NewMockRoleQuerier(t)
		mockAdmin(mockQueries)
		service := NewRoleService(mockQueries, newTestCache())
		got, err := service.Can(ctx, actorID, ActionManageProject, Resource{ProjectID: "PR01"})
		require.NoError(t, err)
		assert.Equal(t, service.CanManageProject(ctx, actorID, "PR01"), got)
		assert.True(t, got)
	})

	t.Run("manage_church matches CanManageChurch", func(t *testing.T) {
		mockQueries := mocks.NewMockRoleQuerier(t)
		mockAdmin(mockQueries)
		service := NewRoleService(mockQueries, newTestCache())
		got, err := service.Can(ctx, actorID, ActionManageChurch, Resource{ChurchID: "CH01"})
		require.NoError(t, err)
		assert.Equal(t, service.CanManageChurch(ctx, actorID, "CH01"), got)
		assert.True(t, got)
	})

	t.Run("manage_team matches CanManageTeam", func(t *testing.T) {
		mockQueries := mocks.NewMockRoleQuerier(t)
		mockAdmin(mockQueries)
		service := NewRoleService(mockQueries, newTestCache())
		got, err := service.Can(ctx, actorID, ActionManageTeam, Resource{TeamID: "TM01"})
		require.NoError(t, err)
		assert.Equal(t, service.CanManageTeam(ctx, actorID, "TM01"), got)
		assert.True(t, got)
	})

	t.Run("assign_role matches CanAssignRole", func(t *testing.T) {
		mockQueries := mocks.NewMockRoleQuerier(t)
		mockAdmin(mockQueries)
		service := NewRoleService(mockQueries, newTestCache())
		churchID := "CH01"
		got, err := service.Can(ctx, actorID, ActionAssignRole, Resource{Role: RoleChurchAdmin, ChurchScope: &churchID})
		require.NoError(t, err)
		want, wantErr := service.CanAssignRole(ctx, actorID, RoleChurchAdmin, &churchID, nil, nil)
		require.NoError(t, wantErr)
		assert.Equal(t, want, got)
		assert.True(t, got)
	})

	t.Run("unknown action errors", func(t *testing.T) {
		mockQueries := mocks.NewMockRoleQuerier(t)
		service := NewRoleService(mockQueries, newTestCache())
		_, err := service.Can(ctx, actorID, Action("bogus"), Resource{})
		require.Error(t, err)
	})
}
