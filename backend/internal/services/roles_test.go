package services

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
)

// newTestCache creates a cache instance for testing
func newTestCache() *cache.CacheWithRegistry {
	c, _ := cache.NewCacheWithRegistry(cache.DefaultConfig())
	return c
}

func TestCanAssignRole_SuperAdminCanAssignAnything(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	service := NewRoleService(mockQueries, newTestCache())

	ctx := context.Background()
	assignerID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// Mock: assigner is a superadmin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: assignerID,
		Role:   string(RoleSuperAdmin),
	}).Return(true, nil)

	// Test assigning all role types
	roles := []RoleType{RoleSuperAdmin, RoleAdmin, RoleChurchAdmin, RoleProjectAdmin, RoleTeamLead, RoleUser}

	for _, role := range roles {
		canAssign, err := service.CanAssignRole(ctx, assignerID, role, nil, nil, nil)
		assert.NoError(t, err)
		assert.True(t, canAssign, "Superadmin should be able to assign %s", role)
	}

}

func TestCanAssignRole_AdminCanAssignLimitedRoles(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	service := NewRoleService(mockQueries, newTestCache())

	ctx := context.Background()
	assignerID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// Mock: assigner is not a superadmin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: assignerID,
		Role:   string(RoleSuperAdmin),
	}).Return(false, nil)

	// Mock: assigner is an admin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: assignerID,
		Role:   string(RoleAdmin),
	}).Return(true, nil)

	// Admin CAN assign these roles
	allowedRoles := []RoleType{RoleChurchAdmin, RoleProjectAdmin, RoleTeamLead}
	for _, role := range allowedRoles {
		canAssign, err := service.CanAssignRole(ctx, assignerID, role, nil, nil, nil)
		assert.NoError(t, err)
		assert.True(t, canAssign, "Admin should be able to assign %s", role)
	}

	// Admin CANNOT assign these roles
	disallowedRoles := []RoleType{RoleSuperAdmin, RoleAdmin, RoleUser}
	for _, role := range disallowedRoles {
		canAssign, err := service.CanAssignRole(ctx, assignerID, role, nil, nil, nil)
		assert.NoError(t, err)
		assert.False(t, canAssign, "Admin should NOT be able to assign %s", role)
	}

}

func TestCanAssignRole_ChurchAdminCanAssignTeamLead(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	service := NewRoleService(mockQueries, newTestCache())

	ctx := context.Background()
	assignerID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	teamID := "TM01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// Mock: assigner is not a superadmin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: assignerID,
		Role:   string(RoleSuperAdmin),
	}).Return(false, nil)

	// Mock: assigner is not an admin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: assignerID,
		Role:   string(RoleAdmin),
	}).Return(false, nil)

	// Mock: assigner has church admin role
	mockQueries.On("GetUserRoles", ctx, assignerID).Return([]*sqlc.UserRole{
		{
			ID:       "UR01ARZ3NDEKTSV4RRFFQ69G5FAV",
			UserID:   assignerID,
			Role:     string(RoleChurchAdmin),
			ChurchID: stringPtr("CH01ARZ3NDEKTSV4RRFFQ69G5FAV"),
		},
	}, nil)

	// Church admin CAN assign team lead
	canAssign, err := service.CanAssignRole(ctx, assignerID, RoleTeamLead, nil, nil, &teamID)
	assert.NoError(t, err)
	assert.True(t, canAssign, "Church admin should be able to assign team lead")

	// Church admin CANNOT assign other roles
	canAssign, err = service.CanAssignRole(ctx, assignerID, RoleProjectAdmin, nil, nil, nil)
	assert.NoError(t, err)
	assert.False(t, canAssign, "Church admin should NOT be able to assign project admin")

}

func TestCanAssignRole_ChurchAdminCanAssignChurchAdminInOwnChurch(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	service := NewRoleService(mockQueries, newTestCache())

	ctx := context.Background()
	assignerID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	churchID := "CH01ARZ3NDEKTSV4RRFFQ69G5FAV"
	otherChurchID := "CH02ARZ3NDEKTSV4RRFFQ69G5FAV"

	// Mock: assigner is not a superadmin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: assignerID,
		Role:   string(RoleSuperAdmin),
	}).Return(false, nil)

	// Mock: assigner is not an admin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: assignerID,
		Role:   string(RoleAdmin),
	}).Return(false, nil)

	// Mock: assigner is church admin for their church
	mockQueries.On("HasRoleInChurch", ctx, sqlc.HasRoleInChurchParams{
		UserID:   assignerID,
		Role:     string(RoleChurchAdmin),
		ChurchID: &churchID,
	}).Return(true, nil)

	// Mock: assigner is NOT church admin for another church
	mockQueries.On("HasRoleInChurch", ctx, sqlc.HasRoleInChurchParams{
		UserID:   assignerID,
		Role:     string(RoleChurchAdmin),
		ChurchID: &otherChurchID,
	}).Return(false, nil)

	// Church admin CAN assign church admin role in their own church
	canAssign, err := service.CanAssignRole(ctx, assignerID, RoleChurchAdmin, &churchID, nil, nil)
	assert.NoError(t, err)
	assert.True(t, canAssign, "Church admin should be able to assign church admin in their own church")

	// Church admin CANNOT assign church admin role in another church
	canAssign, err = service.CanAssignRole(ctx, assignerID, RoleChurchAdmin, &otherChurchID, nil, nil)
	assert.NoError(t, err)
	assert.False(t, canAssign, "Church admin should NOT be able to assign church admin in another church")

}

func TestIsAdmin_ReturnsTrueForSuperAdmin(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	service := NewRoleService(mockQueries, newTestCache())

	ctx := context.Background()
	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// User is a superadmin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: userID,
		Role:   string(RoleSuperAdmin),
	}).Return(true, nil)

	isAdmin := service.IsAdmin(ctx, userID)
	assert.True(t, isAdmin)

}

func TestIsAdmin_ReturnsTrueForAdmin(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	service := NewRoleService(mockQueries, newTestCache())

	ctx := context.Background()
	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// User is not a superadmin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: userID,
		Role:   string(RoleSuperAdmin),
	}).Return(false, nil)

	// User is an admin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: userID,
		Role:   string(RoleAdmin),
	}).Return(true, nil)

	isAdmin := service.IsAdmin(ctx, userID)
	assert.True(t, isAdmin)

}

func TestCanManageProject_AdminCanManageAnyProject(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	service := NewRoleService(mockQueries, newTestCache())

	ctx := context.Background()
	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// User is not a superadmin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: userID,
		Role:   string(RoleSuperAdmin),
	}).Return(false, nil)

	// User is an admin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: userID,
		Role:   string(RoleAdmin),
	}).Return(true, nil)

	canManage := service.CanManageProject(ctx, userID, projectID)
	assert.True(t, canManage)

}

func TestCanManageProject_ProjectAdminCanManageTheirProject(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	service := NewRoleService(mockQueries, newTestCache())

	ctx := context.Background()
	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// User is not a superadmin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: userID,
		Role:   string(RoleSuperAdmin),
	}).Return(false, nil)

	// User is not an admin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: userID,
		Role:   string(RoleAdmin),
	}).Return(false, nil)

	// User is a project admin for this project
	mockQueries.On("HasRoleInProject", ctx, sqlc.HasRoleInProjectParams{
		UserID:    userID,
		Role:      string(RoleProjectAdmin),
		ProjectID: &projectID,
	}).Return(true, nil)

	canManage := service.CanManageProject(ctx, userID, projectID)
	assert.True(t, canManage)

}

func TestAssignRole_WithAuthorization(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	service := NewRoleService(mockQueries, newTestCache())

	ctx := context.Background()
	assignerID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	targetUserID := "US02ARZ3NDEKTSV4RRFFQ69G5FAV"
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// Mock: assigner is not a superadmin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: assignerID,
		Role:   string(RoleSuperAdmin),
	}).Return(false, nil)

	// Mock: assigner is an admin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: assignerID,
		Role:   string(RoleAdmin),
	}).Return(true, nil)

	// Mock: role assignment succeeds
	now := time.Now()
	expectedRole := &sqlc.UserRole{
		ID:         "UR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		UserID:     targetUserID,
		Role:       string(RoleProjectAdmin),
		ProjectID:  &projectID,
		AssignedBy: assignerID,
		AssignedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	mockQueries.On("AssignRole", ctx, mock.MatchedBy(func(params sqlc.AssignRoleParams) bool {
		return params.UserID == targetUserID &&
			params.Role == string(RoleProjectAdmin) &&
			params.ProjectID != nil && *params.ProjectID == projectID &&
			params.AssignedBy == assignerID
	})).Return(expectedRole, nil)

	// Assign the role
	result, err := service.AssignRole(ctx, assignerID, targetUserID, RoleProjectAdmin, nil, &projectID, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, targetUserID, result.UserID)
	assert.Equal(t, string(RoleProjectAdmin), result.Role)

}

func TestAssignRole_UnauthorizedFails(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	service := NewRoleService(mockQueries, newTestCache())

	ctx := context.Background()
	assignerID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	targetUserID := "US02ARZ3NDEKTSV4RRFFQ69G5FAV"

	// Mock: assigner is not a superadmin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: assignerID,
		Role:   string(RoleSuperAdmin),
	}).Return(false, nil)

	// Mock: assigner is not an admin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: assignerID,
		Role:   string(RoleAdmin),
	}).Return(false, nil)

	// Try to assign admin role (should fail - no need to mock GetUserRoles as it won't be called for RoleAdmin)
	result, err := service.AssignRole(ctx, assignerID, targetUserID, RoleAdmin, nil, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "does not have permission")

}

func TestRevokeRole_WithAuthorization(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	service := NewRoleService(mockQueries, newTestCache())

	ctx := context.Background()
	revokerID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	targetUserID := "US02ARZ3NDEKTSV4RRFFQ69G5FAV"
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// Mock: revoker is not a superadmin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: revokerID,
		Role:   string(RoleSuperAdmin),
	}).Return(false, nil)

	// Mock: revoker is an admin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: revokerID,
		Role:   string(RoleAdmin),
	}).Return(true, nil)

	// Mock: role revocation succeeds
	mockQueries.On("RevokeRole", ctx, mock.MatchedBy(func(params sqlc.RevokeRoleParams) bool {
		return params.UserID == targetUserID &&
			params.Role == string(RoleProjectAdmin) &&
			params.ProjectID != nil && *params.ProjectID == projectID
	})).Return(nil)

	// Revoke the role
	err := service.RevokeRole(ctx, revokerID, targetUserID, RoleProjectAdmin, nil, &projectID, nil)

	assert.NoError(t, err)

}

func TestRevokeRole_UnauthorizedFails(t *testing.T) {
	mockQueries := mocks.NewMockRoleQuerier(t)
	service := NewRoleService(mockQueries, newTestCache())

	ctx := context.Background()
	revokerID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
	targetUserID := "US02ARZ3NDEKTSV4RRFFQ69G5FAV"

	// Mock: revoker is not a superadmin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: revokerID,
		Role:   string(RoleSuperAdmin),
	}).Return(false, nil)

	// Mock: revoker is not an admin
	mockQueries.On("HasRole", ctx, sqlc.HasRoleParams{
		UserID: revokerID,
		Role:   string(RoleAdmin),
	}).Return(false, nil)

	// Try to revoke admin role (should fail)
	err := service.RevokeRole(ctx, revokerID, targetUserID, RoleAdmin, nil, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not have permission")

}

// Helper function
func stringPtr(s string) *string {
	return &s
}
