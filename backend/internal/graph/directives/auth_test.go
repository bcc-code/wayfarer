package directives

import (
	"context"
	"fmt"
	"testing"

	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		contextRole    interface{} // Can be string, nil, or other type
		allowedRoles   []string
		shouldCallNext bool
		wantErr        bool
		expectedError  string
	}{
		{
			name:           "user with matching role",
			contextRole:    "admin",
			allowedRoles:   []string{"admin"},
			shouldCallNext: true,
			wantErr:        false,
		},
		{
			name:           "user with one of multiple allowed roles",
			contextRole:    "user",
			allowedRoles:   []string{"admin", "user", "m2m"},
			shouldCallNext: true,
			wantErr:        false,
		},
		{
			name:           "user with first allowed role",
			contextRole:    "admin",
			allowedRoles:   []string{"admin", "user"},
			shouldCallNext: true,
			wantErr:        false,
		},
		{
			name:           "user with last allowed role",
			contextRole:    "m2m",
			allowedRoles:   []string{"admin", "user", "m2m"},
			shouldCallNext: true,
			wantErr:        false,
		},
		{
			name:           "user with mismatched role",
			contextRole:    "user",
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: role 'user' is not allowed",
		},
		{
			name:           "no role in context",
			contextRole:    nil,
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: no role found in context",
		},
		{
			name:           "empty role string",
			contextRole:    "",
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: no role found in context",
		},
		{
			name:           "wrong type in context - int",
			contextRole:    123,
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: no role found in context",
		},
		{
			name:           "wrong type in context - bool",
			contextRole:    true,
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: no role found in context",
		},
		{
			name:           "empty allowed roles list",
			contextRole:    "admin",
			allowedRoles:   []string{},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: role 'admin' is not allowed",
		},
		{
			name:           "case sensitive role matching - wrong case",
			contextRole:    "Admin",
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: role 'Admin' is not allowed",
		},
		{
			name:           "case sensitive role matching - uppercase",
			contextRole:    "ADMIN",
			allowedRoles:   []string{"admin"},
			shouldCallNext: false,
			wantErr:        true,
			expectedError:  "unauthorized: role 'ADMIN' is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup context
			ctx := context.Background()
			if tt.contextRole != nil {
				ctx = context.WithValue(ctx, middleware.UserRoleKey, tt.contextRole)
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
	ctx := context.WithValue(context.Background(), middleware.UserRoleKey, "admin")

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
			ctx := context.WithValue(context.Background(), middleware.UserRoleKey, "user")

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
			ctx := context.WithValue(context.Background(), middleware.UserRoleKey, "admin")

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
