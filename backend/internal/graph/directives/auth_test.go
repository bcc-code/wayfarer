package directives

import (
	"context"
	"fmt"
	"testing"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		contextRoles   interface{} // Can be []string, nil, or other type
		allowedRoles   []string
		shouldCallNext bool
		wantErr        bool
		expectedError  string
	}{
		{
			name:           "user with matching role",
			contextRoles:   []string{"admin"},
			allowedRoles:   []string{"admin"},
			shouldCallNext: true,
			wantErr:        false,
		},
		{
			name:           "user with one of multiple allowed roles",
			contextRoles:   []string{"user"},
			allowedRoles:   []string{"admin", "user", "m2m"},
			shouldCallNext: true,
			wantErr:        false,
		},
		{
			name:           "user with first allowed role",
			contextRoles:   []string{"admin"},
			allowedRoles:   []string{"admin", "user"},
			shouldCallNext: true,
			wantErr:        false,
		},
		{
			name:           "user with last allowed role",
			contextRoles:   []string{"m2m"},
			allowedRoles:   []string{"admin", "user", "m2m"},
			shouldCallNext: true,
			wantErr:        false,
		},
		{
			name:           "user with multiple roles, one matching",
			contextRoles:   []string{"user", "admin"},
			allowedRoles:   []string{"admin"},
			shouldCallNext: true,
			wantErr:        false,
		},
		{
			name:           "user with mismatched role",
			contextRoles:   []string{"user"},
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: user roles",
		},
		{
			name:           "no role in context",
			contextRoles:   nil,
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: no roles found in context",
		},
		{
			name:           "empty role array",
			contextRoles:   []string{},
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: no roles found in context",
		},
		{
			name:           "wrong type in context - int",
			contextRoles:   123,
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: no roles found in context",
		},
		{
			name:           "wrong type in context - bool",
			contextRoles:   true,
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: no roles found in context",
		},
		{
			name:           "wrong type in context - string",
			contextRoles:   "admin",
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: no roles found in context",
		},
		{
			name:           "empty allowed roles list",
			contextRoles:   []string{"admin"},
			allowedRoles:   []string{},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: user roles",
		},
		{
			name:           "case sensitive role matching - wrong case",
			contextRoles:   []string{"Admin"},
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: user roles",
		},
		{
			name:           "case sensitive role matching - uppercase",
			contextRoles:   []string{"ADMIN"},
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: user roles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup context
			ctx := context.Background()
			if tt.contextRoles != nil {
				ctx = context.WithValue(ctx, middleware.UserRolesKey, tt.contextRoles)
			}

			// Track if next was called
			nextCalled := false
			mockNext := func(ctx context.Context) (interface{}, error) {
				nextCalled = true
				return "resolver-result", nil
			}

			// Execute
			result, err := RequireRole(ctx, nil, mockNext, tt.allowedRoles)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "resolver-result", result)
			}

			assert.Equal(t, tt.shouldCallNext, nextCalled, "next resolver call mismatch")
		})
	}
}

func TestRequireRole_NextResolverError(t *testing.T) {
	ctx := context.WithValue(context.Background(), middleware.UserRolesKey, []string{"admin"})

	expectedError := fmt.Errorf("resolver error")
	mockNext := func(ctx context.Context) (interface{}, error) {
		return nil, expectedError
	}

	result, err := RequireRole(ctx, nil, mockNext, []string{"admin"})

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestRequireRole_NextResolverReturnsValue(t *testing.T) {
	tests := []struct {
		name           string
		expectedResult interface{}
	}{
		{
			name:           "string result",
			expectedResult: "test-result",
		},
		{
			name:           "map result",
			expectedResult: map[string]string{"data": "test"},
		},
		{
			name:           "struct result",
			expectedResult: struct{ ID string }{ID: "123"},
		},
		{
			name:           "nil result",
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), middleware.UserRolesKey, []string{"user"})

			mockNext := func(ctx context.Context) (interface{}, error) {
				return tt.expectedResult, nil
			}

			result, err := RequireRole(ctx, nil, mockNext, []string{"user"})

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestRequireRole_NextResolverMultipleErrors(t *testing.T) {
	tests := []struct {
		name          string
		resolverError error
	}{
		{
			name:          "simple error",
			resolverError: fmt.Errorf("simple error"),
		},
		{
			name:          "wrapped error",
			resolverError: fmt.Errorf("wrapped: %w", fmt.Errorf("inner error")),
		},
		{
			name:          "nil error with nil result",
			resolverError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), middleware.UserRolesKey, []string{"admin"})

			mockNext := func(ctx context.Context) (interface{}, error) {
				if tt.resolverError != nil {
					return nil, tt.resolverError
				}
				return nil, nil
			}

			result, err := RequireRole(ctx, nil, mockNext, []string{"admin"})

			if tt.resolverError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.resolverError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, result)
			}
		})
	}
}

// createMockRolesLoader creates a mock dataloader for testing NewRequireRole
func createMockRolesLoader(batchFunc func(context.Context, []string) []*dataloader.Result[[]*model.UserRole]) *dataloader.Loader[string, []*model.UserRole] {
	return dataloader.NewBatchedLoader(
		batchFunc,
		dataloader.WithBatchCapacity[string, []*model.UserRole](100),
		dataloader.WithCache[string, []*model.UserRole](&dataloader.NoCache[string, []*model.UserRole]{}),
	)
}

func TestNewRequireRole_M2MUsesTokenRoles(t *testing.T) {
	// M2M users should use token roles, dataloader should not be called
	loaderCalled := false
	mockLoader := createMockRolesLoader(func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.UserRole] {
		loaderCalled = true
		t.Error("dataloader should not be called for M2M users")
		return nil
	})

	requireRole := NewRequireRole(mockLoader)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "US12345")
	ctx = context.WithValue(ctx, middleware.UserRolesKey, []string{"m2m", "admin"})

	nextCalled := false
	mockNext := func(ctx context.Context) (interface{}, error) {
		nextCalled = true
		return "result", nil
	}

	result, err := requireRole(ctx, nil, mockNext, []string{"admin"})

	require.NoError(t, err)
	assert.True(t, nextCalled, "next resolver should be called")
	assert.Equal(t, "result", result)
	assert.False(t, loaderCalled, "dataloader should not be called for M2M users")
}

func TestNewRequireRole_M2MWithMismatchedRoles(t *testing.T) {
	mockLoader := createMockRolesLoader(func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.UserRole] {
		t.Error("dataloader should not be called for M2M users")
		return nil
	})

	requireRole := NewRequireRole(mockLoader)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "US12345")
	ctx = context.WithValue(ctx, middleware.UserRolesKey, []string{"m2m", "user"})

	mockNext := func(ctx context.Context) (interface{}, error) {
		return "result", nil
	}

	_, err := requireRole(ctx, nil, mockNext, []string{"admin"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized: M2M roles")
}

func TestNewRequireRole_NonM2MUsesDBRoles(t *testing.T) {
	// Non-M2M users should lookup roles from dataloader
	loaderCalled := false
	mockLoader := createMockRolesLoader(func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.UserRole] {
		loaderCalled = true
		results := make([]*dataloader.Result[[]*model.UserRole], len(userIDs))
		for i := range userIDs {
			results[i] = &dataloader.Result[[]*model.UserRole]{
				Data: []*model.UserRole{{Role: model.RoleTypeAdmin}},
			}
		}
		return results
	})

	requireRole := NewRequireRole(mockLoader)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "US12345")
	ctx = context.WithValue(ctx, middleware.UserRolesKey, []string{"user"}) // Token has user, DB has admin

	nextCalled := false
	mockNext := func(ctx context.Context) (interface{}, error) {
		nextCalled = true
		return "result", nil
	}

	result, err := requireRole(ctx, nil, mockNext, []string{"admin"})

	require.NoError(t, err)
	assert.True(t, nextCalled)
	assert.Equal(t, "result", result)
	assert.True(t, loaderCalled, "dataloader should be called for non-M2M users")
}

func TestNewRequireRole_NonM2MWithNoDBRoles_DeniedForAdmin(t *testing.T) {
	// User with no roles in database should be denied admin access
	// but they still have implicit "user" role
	mockLoader := createMockRolesLoader(func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.UserRole] {
		results := make([]*dataloader.Result[[]*model.UserRole], len(userIDs))
		for i := range userIDs {
			results[i] = &dataloader.Result[[]*model.UserRole]{
				Data: []*model.UserRole{}, // Empty roles
			}
		}
		return results
	})

	requireRole := NewRequireRole(mockLoader)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "US12345")
	ctx = context.WithValue(ctx, middleware.UserRolesKey, []string{"user"})

	mockNext := func(ctx context.Context) (interface{}, error) {
		return "result", nil
	}

	_, err := requireRole(ctx, nil, mockNext, []string{"admin"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized: user roles")
}

func TestNewRequireRole_NonM2MWithNoDBRoles_AllowedForUser(t *testing.T) {
	// User with no roles in database should still have implicit "user" role
	// and be allowed to access user-level endpoints
	mockLoader := createMockRolesLoader(func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.UserRole] {
		results := make([]*dataloader.Result[[]*model.UserRole], len(userIDs))
		for i := range userIDs {
			results[i] = &dataloader.Result[[]*model.UserRole]{
				Data: []*model.UserRole{}, // Empty roles
			}
		}
		return results
	})

	requireRole := NewRequireRole(mockLoader)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "US12345")
	ctx = context.WithValue(ctx, middleware.UserRolesKey, []string{"user"})

	nextCalled := false
	mockNext := func(ctx context.Context) (interface{}, error) {
		nextCalled = true
		return "result", nil
	}

	result, err := requireRole(ctx, nil, mockNext, []string{"user"})

	require.NoError(t, err)
	assert.True(t, nextCalled)
	assert.Equal(t, "result", result)
}

func TestNewRequireRole_DataloaderError(t *testing.T) {
	expectedErr := fmt.Errorf("database connection failed")
	mockLoader := createMockRolesLoader(func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.UserRole] {
		results := make([]*dataloader.Result[[]*model.UserRole], len(userIDs))
		for i := range userIDs {
			results[i] = &dataloader.Result[[]*model.UserRole]{
				Error: expectedErr,
			}
		}
		return results
	})

	requireRole := NewRequireRole(mockLoader)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "US12345")
	ctx = context.WithValue(ctx, middleware.UserRolesKey, []string{"user"})

	mockNext := func(ctx context.Context) (interface{}, error) {
		return "result", nil
	}

	_, err := requireRole(ctx, nil, mockNext, []string{"admin"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized: failed to load user roles")
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestNewRequireRole_NoUserID(t *testing.T) {
	mockLoader := createMockRolesLoader(func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.UserRole] {
		return nil
	})

	requireRole := NewRequireRole(mockLoader)

	ctx := context.Background()
	// No user ID in context
	ctx = context.WithValue(ctx, middleware.UserRolesKey, []string{"user"})

	mockNext := func(ctx context.Context) (interface{}, error) {
		return "result", nil
	}

	_, err := requireRole(ctx, nil, mockNext, []string{"admin"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized: user ID not found in context")
}

func TestNewRequireRole_NoTokenRoles(t *testing.T) {
	mockLoader := createMockRolesLoader(func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.UserRole] {
		return nil
	})

	requireRole := NewRequireRole(mockLoader)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "US12345")
	// No roles in context

	mockNext := func(ctx context.Context) (interface{}, error) {
		return "result", nil
	}

	_, err := requireRole(ctx, nil, mockNext, []string{"admin"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized: no roles found in context")
}

func TestNewRequireRole_CaseInsensitiveRoleMatching(t *testing.T) {
	// DB roles are stored as UPPERCASE (e.g., RoleTypeAdmin = "ADMIN")
	// but allowedRoles in directive are lowercase (e.g., "admin")
	// The directive should convert DB roles to lowercase for comparison
	mockLoader := createMockRolesLoader(func(ctx context.Context, userIDs []string) []*dataloader.Result[[]*model.UserRole] {
		results := make([]*dataloader.Result[[]*model.UserRole], len(userIDs))
		for i := range userIDs {
			results[i] = &dataloader.Result[[]*model.UserRole]{
				Data: []*model.UserRole{{Role: model.RoleTypeAdmin}}, // This is "ADMIN"
			}
		}
		return results
	})

	requireRole := NewRequireRole(mockLoader)

	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "US12345")
	ctx = context.WithValue(ctx, middleware.UserRolesKey, []string{"user"})

	nextCalled := false
	mockNext := func(ctx context.Context) (interface{}, error) {
		nextCalled = true
		return "result", nil
	}

	result, err := requireRole(ctx, nil, mockNext, []string{"admin"}) // lowercase "admin"

	require.NoError(t, err)
	assert.True(t, nextCalled)
	assert.Equal(t, "result", result)
}
