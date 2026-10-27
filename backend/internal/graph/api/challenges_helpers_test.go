package api

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
)

// mockBulkEnrollmentRoles is a mock implementation of bulkEnrollmentRoles
type mockBulkEnrollmentRoles struct {
	mock.Mock
}

func (m *mockBulkEnrollmentRoles) CanManageProject(ctx context.Context, userID, projectID string) bool {
	return m.Called(ctx, userID, projectID).Bool(0)
}

func (m *mockBulkEnrollmentRoles) LoadUserRoles(ctx context.Context, userID string) ([]*sqlc.UserRole, error) {
	args := m.Called(ctx, userID)
	roles, _ := args.Get(0).([]*sqlc.UserRole)
	return roles, args.Error(1)
}

// mockChurchAdminScopeQuerier is a mock implementation of churchAdminScopeQuerier
type mockChurchAdminScopeQuerier struct {
	mock.Mock
}

func (m *mockChurchAdminScopeQuerier) GetUsersOutsideChurchAdminScope(ctx context.Context, arg sqlc.GetUsersOutsideChurchAdminScopeParams) ([]string, error) {
	args := m.Called(ctx, arg)
	ids, _ := args.Get(0).([]string)
	return ids, args.Error(1)
}

func TestAuthorizeBulkEnrollment(t *testing.T) {
	ctx := context.Background()
	const (
		callerID  = "US01ARZ3NDEKTSV4RRFFQ69G5FAV"
		projectID = "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
		churchID  = "CH01ARZ3NDEKTSV4RRFFQ69G5FAV"
	)
	userIDs := []string{"US01ARZ3NDEKTSV4RRFFQ69G5FA1", "US01ARZ3NDEKTSV4RRFFQ69G5FA2"}
	church := churchID
	churchAdmin := []*sqlc.UserRole{{Role: string(services.RoleChurchAdmin), ChurchID: &church}}
	scopeParams := sqlc.GetUsersOutsideChurchAdminScopeParams{
		Userids: userIDs, Projectid: projectID, Churchids: []string{churchID},
	}

	t.Run("a project manager may enroll anyone without a scope check", func(t *testing.T) {
		roles := &mockBulkEnrollmentRoles{}
		q := &mockChurchAdminScopeQuerier{}
		roles.On("CanManageProject", ctx, callerID, projectID).Return(true)

		require.NoError(t, authorizeBulkEnrollment(ctx, roles, q, callerID, projectID, userIDs))
		q.AssertNotCalled(t, "GetUsersOutsideChurchAdminScope", mock.Anything, mock.Anything)
	})

	t.Run("an m2m token may enroll anyone without a role lookup", func(t *testing.T) {
		roles := &mockBulkEnrollmentRoles{}
		q := &mockChurchAdminScopeQuerier{}
		m2mCtx := context.WithValue(ctx, middleware.UserRolesKey, []string{"m2m"})

		require.NoError(t, authorizeBulkEnrollment(m2mCtx, roles, q, "M2M_SERVICE", projectID, userIDs))
		roles.AssertNotCalled(t, "CanManageProject", mock.Anything, mock.Anything, mock.Anything)
		q.AssertNotCalled(t, "GetUsersOutsideChurchAdminScope", mock.Anything, mock.Anything)
	})

	t.Run("a church admin may enroll users in scope", func(t *testing.T) {
		roles := &mockBulkEnrollmentRoles{}
		q := &mockChurchAdminScopeQuerier{}
		roles.On("CanManageProject", ctx, callerID, projectID).Return(false)
		roles.On("LoadUserRoles", ctx, callerID).Return(churchAdmin, nil)
		q.On("GetUsersOutsideChurchAdminScope", ctx, scopeParams).Return([]string{}, nil)

		require.NoError(t, authorizeBulkEnrollment(ctx, roles, q, callerID, projectID, userIDs))
		q.AssertExpectations(t)
	})

	t.Run("a church admin is rejected when any user is out of scope", func(t *testing.T) {
		roles := &mockBulkEnrollmentRoles{}
		q := &mockChurchAdminScopeQuerier{}
		roles.On("CanManageProject", ctx, callerID, projectID).Return(false)
		roles.On("LoadUserRoles", ctx, callerID).Return(churchAdmin, nil)
		q.On("GetUsersOutsideChurchAdminScope", ctx, scopeParams).Return([]string{userIDs[1]}, nil)

		err := authorizeBulkEnrollment(ctx, roles, q, callerID, projectID, userIDs)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed to enroll 1 user(s) outside your church")
	})

	t.Run("a caller without a church admin role is rejected", func(t *testing.T) {
		roles := &mockBulkEnrollmentRoles{}
		q := &mockChurchAdminScopeQuerier{}
		roles.On("CanManageProject", ctx, callerID, projectID).Return(false)
		roles.On("LoadUserRoles", ctx, callerID).Return([]*sqlc.UserRole{{Role: "TEAM_LEAD"}}, nil)

		err := authorizeBulkEnrollment(ctx, roles, q, callerID, projectID, userIDs)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not authorized")
		q.AssertNotCalled(t, "GetUsersOutsideChurchAdminScope", mock.Anything, mock.Anything)
	})

	t.Run("query errors are returned", func(t *testing.T) {
		roles := &mockBulkEnrollmentRoles{}
		q := &mockChurchAdminScopeQuerier{}
		roles.On("CanManageProject", ctx, callerID, projectID).Return(false)
		roles.On("LoadUserRoles", ctx, callerID).Return(churchAdmin, nil)
		dbErr := errors.New("connection reset")
		q.On("GetUsersOutsideChurchAdminScope", ctx, scopeParams).Return(nil, dbErr)

		require.ErrorIs(t, authorizeBulkEnrollment(ctx, roles, q, callerID, projectID, userIDs), dbErr)
	})
}
