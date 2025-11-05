package api

import (
	"context"
	"errors"
	"testing"

	"github.com/graph-gophers/dataloader/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// mockUserLoader is a mock implementation of UserLoaderInterface
type mockUserLoader struct {
	mock.Mock
}

func (m *mockUserLoader) Load(ctx context.Context, key string) dataloader.Thunk[*model.User] {
	args := m.Called(ctx, key)
	return args.Get(0).(dataloader.Thunk[*model.User])
}

// mockRoleService is a mock implementation of RoleServiceInterface
type mockRoleService struct {
	mock.Mock
}

func (m *mockRoleService) IsAdmin(ctx context.Context, userID string) bool {
	args := m.Called(ctx, userID)
	return args.Bool(0)
}

func (m *mockRoleService) CanManageChurch(ctx context.Context, userID, churchID string) bool {
	args := m.Called(ctx, userID, churchID)
	return args.Bool(0)
}

func (m *mockRoleService) CanManageProject(ctx context.Context, userID, projectID string) bool {
	args := m.Called(ctx, userID, projectID)
	return args.Bool(0)
}

func (m *mockRoleService) CanManageTeam(ctx context.Context, userID, teamID string) bool {
	args := m.Called(ctx, userID, teamID)
	return args.Bool(0)
}

// TestValidateUserAccess tests the validateUserAccess function
func TestValidateUserAccess(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		setupLoader func() *mockUserLoader
		expectError bool
		errorMsg    string
		expectedID  string
	}{
		{
			name: "valid authenticated user",
			ctx:  createTestContext("US123"),
			setupLoader: func() *mockUserLoader {
				loader := new(mockUserLoader)
				loader.On("Load", mock.Anything, "US123").Return(
					dataloader.Thunk[*model.User](func() (*model.User, error) {
						return &model.User{
							ID:       "US123",
							ChurchID: "CH456",
						}, nil
					}),
				)
				return loader
			},
			expectError: false,
			expectedID:  "US123",
		},
		{
			name:        "missing user ID in context",
			ctx:         createTestContext(""),
			setupLoader: func() *mockUserLoader { return new(mockUserLoader) },
			expectError: true,
			errorMsg:    "user not authenticated",
		},
		{
			name: "thunk returns error when called",
			ctx:  createTestContext("US123"),
			setupLoader: func() *mockUserLoader {
				loader := new(mockUserLoader)
				loader.On("Load", mock.Anything, "US123").Return(
					dataloader.Thunk[*model.User](func() (*model.User, error) {
						return nil, errors.New("user not found")
					}),
				)
				return loader
			},
			expectError: true,
			errorMsg:    "failed to load current user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := tt.setupLoader()
			result, err := validateUserAccess(tt.ctx, loader)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedID, result.UserID)
				assert.Equal(t, "CH456", result.ChurchID)
			}

			loader.AssertExpectations(t)
		})
	}
}

// TestCheckUserPermissions tests the checkUserPermissions function
func TestCheckUserPermissions(t *testing.T) {
	tests := []struct {
		name             string
		userInfo         *authenticatedUserInfo
		filter           *model.UserFilter
		setupRoleService func() *mockRoleService
		expectError      bool
		errorMsg         string
		checkPerms       func(*testing.T, *permissionContext)
	}{
		{
			name: "superadmin has all permissions",
			userInfo: &authenticatedUserInfo{
				UserID:   "US123",
				ChurchID: "CH456",
			},
			filter: nil,
			setupRoleService: func() *mockRoleService {
				svc := new(mockRoleService)
				svc.On("IsAdmin", mock.Anything, "US123").Return(true)
				return svc
			},
			expectError: false,
			checkPerms: func(t *testing.T, perms *permissionContext) {
				assert.True(t, perms.IsAdmin)
				assert.False(t, perms.IsChurchAdmin)
				assert.False(t, perms.IsProjectAdmin)
				assert.False(t, perms.IsTeamLead)
			},
		},
		{
			name: "church admin has church permissions",
			userInfo: &authenticatedUserInfo{
				UserID:   "US123",
				ChurchID: "CH456",
			},
			filter: nil,
			setupRoleService: func() *mockRoleService {
				svc := new(mockRoleService)
				svc.On("IsAdmin", mock.Anything, "US123").Return(false)
				svc.On("CanManageChurch", mock.Anything, "US123", "CH456").Return(true)
				return svc
			},
			expectError: false,
			checkPerms: func(t *testing.T, perms *permissionContext) {
				assert.False(t, perms.IsAdmin)
				assert.True(t, perms.IsChurchAdmin)
				assert.Equal(t, "CH456", perms.ChurchID)
			},
		},
		{
			name: "project admin has project permissions",
			userInfo: &authenticatedUserInfo{
				UserID:   "US123",
				ChurchID: "CH456",
			},
			filter: &model.UserFilter{
				ProjectID: stringPtr("PR789"),
			},
			setupRoleService: func() *mockRoleService {
				svc := new(mockRoleService)
				svc.On("IsAdmin", mock.Anything, "US123").Return(false)
				svc.On("CanManageChurch", mock.Anything, "US123", "CH456").Return(false)
				svc.On("CanManageProject", mock.Anything, "US123", "PR789").Return(true)
				return svc
			},
			expectError: false,
			checkPerms: func(t *testing.T, perms *permissionContext) {
				assert.False(t, perms.IsAdmin)
				assert.False(t, perms.IsChurchAdmin)
				assert.True(t, perms.IsProjectAdmin)
				assert.Equal(t, "PR789", perms.ProjectID)
			},
		},
		{
			name: "team lead has team permissions",
			userInfo: &authenticatedUserInfo{
				UserID:   "US123",
				ChurchID: "CH456",
			},
			filter: &model.UserFilter{
				TeamID: stringPtr("TM999"),
			},
			setupRoleService: func() *mockRoleService {
				svc := new(mockRoleService)
				svc.On("IsAdmin", mock.Anything, "US123").Return(false)
				svc.On("CanManageChurch", mock.Anything, "US123", "CH456").Return(false)
				svc.On("CanManageTeam", mock.Anything, "US123", "TM999").Return(true)
				return svc
			},
			expectError: false,
			checkPerms: func(t *testing.T, perms *permissionContext) {
				assert.False(t, perms.IsAdmin)
				assert.False(t, perms.IsChurchAdmin)
				assert.False(t, perms.IsProjectAdmin)
				assert.True(t, perms.IsTeamLead)
				assert.Equal(t, "TM999", perms.TeamID)
			},
		},
		{
			name: "user with no admin role is denied",
			userInfo: &authenticatedUserInfo{
				UserID:   "US123",
				ChurchID: "CH456",
			},
			filter: nil,
			setupRoleService: func() *mockRoleService {
				svc := new(mockRoleService)
				svc.On("IsAdmin", mock.Anything, "US123").Return(false)
				svc.On("CanManageChurch", mock.Anything, "US123", "CH456").Return(false)
				return svc
			},
			expectError: true,
			errorMsg:    "permission denied",
		},
		{
			name: "user with multiple admin roles",
			userInfo: &authenticatedUserInfo{
				UserID:   "US123",
				ChurchID: "CH456",
			},
			filter: &model.UserFilter{
				ProjectID: stringPtr("PR789"),
				TeamID:    stringPtr("TM999"),
			},
			setupRoleService: func() *mockRoleService {
				svc := new(mockRoleService)
				svc.On("IsAdmin", mock.Anything, "US123").Return(false)
				svc.On("CanManageChurch", mock.Anything, "US123", "CH456").Return(true)
				svc.On("CanManageProject", mock.Anything, "US123", "PR789").Return(true)
				svc.On("CanManageTeam", mock.Anything, "US123", "TM999").Return(true)
				return svc
			},
			expectError: false,
			checkPerms: func(t *testing.T, perms *permissionContext) {
				assert.True(t, perms.IsChurchAdmin)
				assert.True(t, perms.IsProjectAdmin)
				assert.True(t, perms.IsTeamLead)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc := tt.setupRoleService()

			result, err := checkUserPermissions(ctx, svc, tt.userInfo, tt.filter)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkPerms != nil {
					tt.checkPerms(t, result)
				}
			}

			svc.AssertExpectations(t)
		})
	}
}

// TestApplyPermissionFilters tests the applyPermissionFilters function
func TestApplyPermissionFilters(t *testing.T) {
	tests := []struct {
		name           string
		inputFilter    *model.UserFilter
		perms          *permissionContext
		checkResult    func(*testing.T, *model.UserFilter)
		expectedFilter *model.UserFilter
	}{
		{
			name:        "superadmin - no filters added",
			inputFilter: &model.UserFilter{},
			perms: &permissionContext{
				IsAdmin:  true,
				ChurchID: "CH456",
			},
			checkResult: func(t *testing.T, filter *model.UserFilter) {
				assert.Nil(t, filter.ChurchID)
				assert.Nil(t, filter.ProjectID)
				assert.Nil(t, filter.TeamID)
			},
		},
		{
			name:        "church admin - church filter added",
			inputFilter: &model.UserFilter{},
			perms: &permissionContext{
				IsAdmin:       false,
				IsChurchAdmin: true,
				ChurchID:      "CH456",
			},
			checkResult: func(t *testing.T, filter *model.UserFilter) {
				require.NotNil(t, filter.ChurchID)
				assert.Equal(t, "CH456", *filter.ChurchID)
			},
		},
		{
			name:        "project admin - project filter added",
			inputFilter: &model.UserFilter{},
			perms: &permissionContext{
				IsAdmin:        false,
				IsProjectAdmin: true,
				ProjectID:      "PR789",
			},
			checkResult: func(t *testing.T, filter *model.UserFilter) {
				require.NotNil(t, filter.ProjectID)
				assert.Equal(t, "PR789", *filter.ProjectID)
			},
		},
		{
			name:        "team lead - team filter added",
			inputFilter: &model.UserFilter{},
			perms: &permissionContext{
				IsAdmin:    false,
				IsTeamLead: true,
				TeamID:     "TM999",
			},
			checkResult: func(t *testing.T, filter *model.UserFilter) {
				require.NotNil(t, filter.TeamID)
				assert.Equal(t, "TM999", *filter.TeamID)
			},
		},
		{
			name: "project admin with existing project filter - not overridden",
			inputFilter: &model.UserFilter{
				ProjectID: stringPtr("PR123"),
			},
			perms: &permissionContext{
				IsAdmin:        false,
				IsProjectAdmin: true,
				ProjectID:      "PR789",
			},
			checkResult: func(t *testing.T, filter *model.UserFilter) {
				require.NotNil(t, filter.ProjectID)
				assert.Equal(t, "PR123", *filter.ProjectID, "Should not override existing filter")
			},
		},
		{
			name:        "nil filter - initialized and church filter added",
			inputFilter: nil,
			perms: &permissionContext{
				IsAdmin:       false,
				IsChurchAdmin: true,
				ChurchID:      "CH456",
			},
			checkResult: func(t *testing.T, filter *model.UserFilter) {
				require.NotNil(t, filter)
				require.NotNil(t, filter.ChurchID)
				assert.Equal(t, "CH456", *filter.ChurchID)
			},
		},
		{
			name:        "project admin takes precedence over church admin",
			inputFilter: &model.UserFilter{},
			perms: &permissionContext{
				IsAdmin:        false,
				IsChurchAdmin:  true,
				IsProjectAdmin: true,
				ChurchID:       "CH456",
				ProjectID:      "PR789",
			},
			checkResult: func(t *testing.T, filter *model.UserFilter) {
				// Church filter should NOT be added when project admin
				assert.Nil(t, filter.ChurchID)
				require.NotNil(t, filter.ProjectID)
				assert.Equal(t, "PR789", *filter.ProjectID)
			},
		},
		{
			name:        "team lead takes precedence over church admin",
			inputFilter: &model.UserFilter{},
			perms: &permissionContext{
				IsAdmin:       false,
				IsChurchAdmin: true,
				IsTeamLead:    true,
				ChurchID:      "CH456",
				TeamID:        "TM999",
			},
			checkResult: func(t *testing.T, filter *model.UserFilter) {
				// Church filter should NOT be added when team lead
				assert.Nil(t, filter.ChurchID)
				require.NotNil(t, filter.TeamID)
				assert.Equal(t, "TM999", *filter.TeamID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyPermissionFilters(tt.inputFilter, tt.perms)
			require.NotNil(t, result)
			tt.checkResult(t, result)
		})
	}
}

// mockQuerier is a minimal mock for the Querier interface
type mockQuerier struct {
	mock.Mock
}

func (m *mockQuerier) GetUsersFiltered(ctx context.Context, params sqlc.GetUsersFilteredParams) ([]sqlc.GetUsersFilteredRow, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.GetUsersFilteredRow), args.Error(1)
}

// TestBuildUserFilterParams tests the buildUserFilterParams function
func TestBuildUserFilterParams(t *testing.T) {
	maleGender := model.GenderMale
	femaleGender := model.GenderFemale

	tests := []struct {
		name   string
		filter *model.UserFilter
		limit  *int
		offset *int
		check  func(*testing.T, sqlc.GetUsersFilteredParams)
	}{
		{
			name: "all filters populated",
			filter: &model.UserFilter{
				ChurchID:  stringPtr("CH123"),
				Gender:    &maleGender,
				MinAge:    intPtr(18),
				MaxAge:    intPtr(30),
				ProjectID: stringPtr("PR456"),
				EventID:   stringPtr("EV789"),
				TeamID:    stringPtr("TM999"),
				Ids:       []string{"US001", "US002"},
			},
			limit:  intPtr(10),
			offset: intPtr(5),
			check: func(t *testing.T, params sqlc.GetUsersFilteredParams) {
				assert.Equal(t, "CH123", params.Churchid)
				assert.Equal(t, "MALE", params.Gender)
				assert.Equal(t, int32(18), params.Minage)
				assert.Equal(t, int32(30), params.Maxage)
				assert.Equal(t, "PR456", params.Projectid)
				assert.Equal(t, "EV789", params.Eventid)
				assert.Equal(t, "TM999", params.Teamid)
				assert.Equal(t, []string{"US001", "US002"}, params.Ids)
				assert.Equal(t, int32(10), params.Querylimit)
				assert.Equal(t, int32(5), params.Queryoffset)
			},
		},
		{
			name: "minimal filter with female gender",
			filter: &model.UserFilter{
				Gender: &femaleGender,
			},
			limit:  nil,
			offset: nil,
			check: func(t *testing.T, params sqlc.GetUsersFilteredParams) {
				assert.Equal(t, "FEMALE", params.Gender)
				assert.Equal(t, "", params.Churchid)
				assert.Equal(t, int32(0), params.Minage)
				assert.Equal(t, int32(0), params.Maxage)
			},
		},
		{
			name: "only age range filter",
			filter: &model.UserFilter{
				MinAge: intPtr(25),
				MaxAge: intPtr(35),
			},
			limit:  nil,
			offset: nil,
			check: func(t *testing.T, params sqlc.GetUsersFilteredParams) {
				assert.Equal(t, int32(25), params.Minage)
				assert.Equal(t, int32(35), params.Maxage)
			},
		},
		{
			name: "only church filter",
			filter: &model.UserFilter{
				ChurchID: stringPtr("CH999"),
			},
			limit:  nil,
			offset: nil,
			check: func(t *testing.T, params sqlc.GetUsersFilteredParams) {
				assert.Equal(t, "CH999", params.Churchid)
			},
		},
		{
			name: "only IDs filter",
			filter: &model.UserFilter{
				Ids: []string{"US100", "US200", "US300"},
			},
			limit:  nil,
			offset: nil,
			check: func(t *testing.T, params sqlc.GetUsersFilteredParams) {
				assert.Equal(t, []string{"US100", "US200", "US300"}, params.Ids)
			},
		},
		{
			name:   "nil filter with pagination",
			filter: &model.UserFilter{},
			limit:  intPtr(20),
			offset: intPtr(10),
			check: func(t *testing.T, params sqlc.GetUsersFilteredParams) {
				assert.Equal(t, int32(20), params.Querylimit)
				assert.Equal(t, int32(10), params.Queryoffset)
			},
		},
		{
			name: "empty IDs array",
			filter: &model.UserFilter{
				Ids: []string{},
			},
			limit:  nil,
			offset: nil,
			check: func(t *testing.T, params sqlc.GetUsersFilteredParams) {
				assert.NotNil(t, params.Ids)
				assert.Empty(t, params.Ids)
			},
		},
		{
			name: "zero values for age",
			filter: &model.UserFilter{
				MinAge: intPtr(0),
				MaxAge: intPtr(0),
			},
			limit:  nil,
			offset: nil,
			check: func(t *testing.T, params sqlc.GetUsersFilteredParams) {
				assert.Equal(t, int32(0), params.Minage)
				assert.Equal(t, int32(0), params.Maxage)
			},
		},
		{
			name:   "zero pagination values",
			filter: &model.UserFilter{},
			limit:  intPtr(0),
			offset: intPtr(0),
			check: func(t *testing.T, params sqlc.GetUsersFilteredParams) {
				assert.Equal(t, int32(0), params.Querylimit)
				assert.Equal(t, int32(0), params.Queryoffset)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildUserFilterParams(tt.filter, tt.limit, tt.offset)
			tt.check(t, result)
		})
	}
}

// TestUsersResolverIntegration tests the full Users resolver flow
func TestUsersResolverIntegration(t *testing.T) {
	tests := []struct {
		name          string
		ctx           context.Context
		filter        *model.UserFilter
		limit         *int
		offset        *int
		setupMocks    func(*mockUserLoader, *mockRoleService, *mockQuerier)
		expectError   bool
		errorMsg      string
		expectedCount int
	}{
		{
			name:   "superadmin can query all users",
			ctx:    createTestContext("US123"),
			filter: &model.UserFilter{},
			limit:  intPtr(10),
			offset: nil,
			setupMocks: func(loader *mockUserLoader, roles *mockRoleService, querier *mockQuerier) {
				// Setup user loader
				loader.On("Load", mock.Anything, "US123").Return(
					dataloader.Thunk[*model.User](func() (*model.User, error) {
						return &model.User{ID: "US123", ChurchID: "CH456"}, nil
					}),
				)

				// Setup role service - superadmin
				roles.On("IsAdmin", mock.Anything, "US123").Return(true)

				// Setup querier - return some users
				querier.On("GetUsersFiltered", mock.Anything, mock.MatchedBy(func(params sqlc.GetUsersFilteredParams) bool {
					return params.Querylimit == 10
				})).Return([]sqlc.GetUsersFilteredRow{
					{ID: "US001", Name: "User 1", ChurchID: "CH456", Gender: "MALE"},
					{ID: "US002", Name: "User 2", ChurchID: "CH456", Gender: "FEMALE"},
				}, nil)
			},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name: "church admin can only see users from their church",
			ctx:  createTestContext("US123"),
			filter: &model.UserFilter{
				Gender: func() *model.Gender { g := model.GenderMale; return &g }(),
			},
			limit:  nil,
			offset: nil,
			setupMocks: func(loader *mockUserLoader, roles *mockRoleService, querier *mockQuerier) {
				loader.On("Load", mock.Anything, "US123").Return(
					dataloader.Thunk[*model.User](func() (*model.User, error) {
						return &model.User{ID: "US123", ChurchID: "CH456"}, nil
					}),
				)

				roles.On("IsAdmin", mock.Anything, "US123").Return(false)
				roles.On("CanManageChurch", mock.Anything, "US123", "CH456").Return(true)

				// Verify that church filter is applied
				querier.On("GetUsersFiltered", mock.Anything, mock.MatchedBy(func(params sqlc.GetUsersFilteredParams) bool {
					return params.Churchid == "CH456" && params.Gender == "MALE"
				})).Return([]sqlc.GetUsersFilteredRow{
					{ID: "US001", Name: "User 1", ChurchID: "CH456", Gender: "MALE"},
				}, nil)
			},
			expectError:   false,
			expectedCount: 1,
		},
		{
			name:   "unauthenticated user is rejected",
			ctx:    createTestContext(""),
			filter: nil,
			limit:  nil,
			offset: nil,
			setupMocks: func(loader *mockUserLoader, roles *mockRoleService, querier *mockQuerier) {
				// No setup needed - validation should fail before any calls
			},
			expectError: true,
			errorMsg:    "user not authenticated",
		},
		{
			name:   "user without admin permissions is rejected",
			ctx:    createTestContext("US123"),
			filter: nil,
			limit:  nil,
			offset: nil,
			setupMocks: func(loader *mockUserLoader, roles *mockRoleService, querier *mockQuerier) {
				loader.On("Load", mock.Anything, "US123").Return(
					dataloader.Thunk[*model.User](func() (*model.User, error) {
						return &model.User{ID: "US123", ChurchID: "CH456"}, nil
					}),
				)

				roles.On("IsAdmin", mock.Anything, "US123").Return(false)
				roles.On("CanManageChurch", mock.Anything, "US123", "CH456").Return(false)
			},
			expectError: true,
			errorMsg:    "permission denied",
		},
		{
			name: "project admin filtering by project",
			ctx:  createTestContext("US123"),
			filter: &model.UserFilter{
				ProjectID: stringPtr("PR789"),
			},
			limit:  nil,
			offset: nil,
			setupMocks: func(loader *mockUserLoader, roles *mockRoleService, querier *mockQuerier) {
				loader.On("Load", mock.Anything, "US123").Return(
					dataloader.Thunk[*model.User](func() (*model.User, error) {
						return &model.User{ID: "US123", ChurchID: "CH456"}, nil
					}),
				)

				roles.On("IsAdmin", mock.Anything, "US123").Return(false)
				roles.On("CanManageChurch", mock.Anything, "US123", "CH456").Return(false)
				roles.On("CanManageProject", mock.Anything, "US123", "PR789").Return(true)

				querier.On("GetUsersFiltered", mock.Anything, mock.MatchedBy(func(params sqlc.GetUsersFilteredParams) bool {
					return params.Projectid == "PR789"
				})).Return([]sqlc.GetUsersFilteredRow{
					{ID: "US001", Name: "User 1", ChurchID: "CH456", Gender: "MALE"},
				}, nil)
			},
			expectError:   false,
			expectedCount: 1,
		},
		{
			name: "database error is propagated",
			ctx:  createTestContext("US123"),
			filter: &model.UserFilter{
				ChurchID: stringPtr("CH456"),
			},
			limit:  nil,
			offset: nil,
			setupMocks: func(loader *mockUserLoader, roles *mockRoleService, querier *mockQuerier) {
				loader.On("Load", mock.Anything, "US123").Return(
					dataloader.Thunk[*model.User](func() (*model.User, error) {
						return &model.User{ID: "US123", ChurchID: "CH456"}, nil
					}),
				)

				roles.On("IsAdmin", mock.Anything, "US123").Return(true)

				querier.On("GetUsersFiltered", mock.Anything, mock.Anything).Return(
					nil, errors.New("database connection error"),
				)
			},
			expectError: true,
			errorMsg:    "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			loader := new(mockUserLoader)
			roles := new(mockRoleService)
			querier := new(mockQuerier)

			// Setup mocks
			tt.setupMocks(loader, roles, querier)

			// Create a mock resolver with minimal setup
			// Note: We can't easily test the full resolver without setting up
			// DB, Loaders, Cache, etc. Instead, we test the integration of
			// our extracted functions which the resolver uses.

			// Test the flow: validate -> check permissions -> apply filters -> build params
			userInfo, err := validateUserAccess(tt.ctx, loader)
			if tt.expectError && err != nil {
				// Expected error at validation stage
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error at validation: %v", err)
			}

			perms, err := checkUserPermissions(tt.ctx, roles, userInfo, tt.filter)
			if tt.expectError && err != nil {
				// Expected error at permission check stage
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error at permission check: %v", err)
			}

			filter := applyPermissionFilters(tt.filter, perms)
			params := buildUserFilterParams(filter, tt.limit, tt.offset)

			// Query database
			rows, err := querier.GetUsersFiltered(tt.ctx, params)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Len(t, rows, tt.expectedCount)
			}

			// Verify all mocks were called as expected
			loader.AssertExpectations(t)
			roles.AssertExpectations(t)
			querier.AssertExpectations(t)
		})
	}
}
