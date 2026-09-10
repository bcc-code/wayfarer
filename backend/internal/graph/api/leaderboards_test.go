package api

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalLeaderboardFilter_Nil(t *testing.T) {
	b, err := marshalLeaderboardFilter(nil)
	require.NoError(t, err)
	assert.Nil(t, b)
}

func TestMarshalLeaderboardFilter_ValidFilter(t *testing.T) {
	minScore := 10
	churchID := "CH01ARZ3NDEKTSV4RRFFQ69G5FAV"
	filter := &model.LeaderboardFilter{
		MinScore: &minScore,
		ChurchID: &churchID,
	}

	b, err := marshalLeaderboardFilter(filter)
	require.NoError(t, err)
	require.NotNil(t, b)
	assert.Contains(t, string(b), `"minScore":10`)
	assert.Contains(t, string(b), churchID)
}

func TestFilterConfigsByVisibility_NonAdminSeesOnlyActive(t *testing.T) {
	configs := []*model.LeaderboardConfig{
		{ID: "LC1", IsActive: true},
		{ID: "LC2", IsActive: false},
		{ID: "LC3", IsActive: true},
	}

	result := filterConfigsByVisibility(configs, false)

	require.Len(t, result, 2)
	assert.Equal(t, "LC1", result[0].ID)
	assert.Equal(t, "LC3", result[1].ID)
}

func TestFilterConfigsByVisibility_AdminSeesAll(t *testing.T) {
	configs := []*model.LeaderboardConfig{
		{ID: "LC1", IsActive: true},
		{ID: "LC2", IsActive: false},
	}

	result := filterConfigsByVisibility(configs, true)

	require.Len(t, result, 2)
}

func TestFilterConfigsByVisibility_EmptyInput(t *testing.T) {
	result := filterConfigsByVisibility(nil, false)
	assert.Empty(t, result)
}

func TestBuildLeaderboardParamsFromConfig_ProjectScoped(t *testing.T) {
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	config := &model.LeaderboardConfig{
		ProjectID:  projectID,
		EventID:    nil,
		EntityType: model.LeaderboardEntityTypePersons,
	}

	params, isEvent, err := buildLeaderboardParamsFromConfig(config, nil, nil, nil, nil, "US01ARZ3NDEKTSV4RRFFQ69G5FAV")

	require.NoError(t, err)
	assert.False(t, isEvent)
	assert.Equal(t, projectID, params.ContextID)
	assert.Equal(t, model.LeaderboardEntityTypePersons, params.EntityType)
	assert.Equal(t, "US01ARZ3NDEKTSV4RRFFQ69G5FAV", params.UserID)
	assert.Nil(t, params.Filter)
}

func TestBuildLeaderboardParamsFromConfig_EventScoped(t *testing.T) {
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	eventID := "EV01ARZ3NDEKTSV4RRFFQ69G5FAV"
	config := &model.LeaderboardConfig{
		ProjectID:  projectID,
		EventID:    &eventID,
		EntityType: model.LeaderboardEntityTypeTeams,
	}

	params, isEvent, err := buildLeaderboardParamsFromConfig(config, nil, nil, nil, nil, "US01ARZ3NDEKTSV4RRFFQ69G5FAV")

	require.NoError(t, err)
	assert.True(t, isEvent)
	assert.Equal(t, eventID, params.ContextID)
}

func TestBuildLeaderboardParamsFromConfig_ParsesFilter(t *testing.T) {
	filterJSON := `{"minScore":42}`
	config := &model.LeaderboardConfig{
		ProjectID:  "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		EntityType: model.LeaderboardEntityTypePersons,
		Filter:     &filterJSON,
	}

	params, _, err := buildLeaderboardParamsFromConfig(config, nil, nil, nil, nil, "US01ARZ3NDEKTSV4RRFFQ69G5FAV")

	require.NoError(t, err)
	require.NotNil(t, params.Filter)
	require.NotNil(t, params.Filter.MinScore)
	assert.Equal(t, 42, *params.Filter.MinScore)
}

func TestBuildLeaderboardParamsFromConfig_InvalidFilterJSON(t *testing.T) {
	invalidJSON := `{not valid json`
	config := &model.LeaderboardConfig{
		ProjectID:  "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
		EntityType: model.LeaderboardEntityTypePersons,
		Filter:     &invalidJSON,
	}

	_, _, err := buildLeaderboardParamsFromConfig(config, nil, nil, nil, nil, "US01ARZ3NDEKTSV4RRFFQ69G5FAV")

	assert.Error(t, err)
}

func TestBuildLeaderboardConfigFilterParamsCursor_AppliesFilterFields(t *testing.T) {
	projectID := "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
	isActive := true
	filter := &model.LeaderboardConfigFilter{
		ProjectID: &projectID,
		IsActive:  &isActive,
	}

	params, err := buildLeaderboardConfigFilterParamsCursor(filter, nil, nil, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, projectID, params.Projectid)
	require.NotNil(t, params.Isactive)
	assert.True(t, *params.Isactive)
	assert.Equal(t, int32(11), params.Querylimit) // default page size + 1
	assert.False(t, params.Isbackward)
}

func TestBuildLeaderboardConfigFilterParamsCursor_FirstAndLastMutuallyExclusive(t *testing.T) {
	first := 5
	last := 5

	_, err := buildLeaderboardConfigFilterParamsCursor(nil, &first, nil, &last, nil)

	assert.Error(t, err)
}

func TestBuildLeaderboardConfigFilterParamsCursor_BackwardPagination(t *testing.T) {
	last := 5

	params, err := buildLeaderboardConfigFilterParamsCursor(nil, nil, nil, &last, nil)

	require.NoError(t, err)
	assert.True(t, params.Isbackward)
	assert.Equal(t, int32(6), params.Querylimit)
}

func TestBuildLeaderboardConfigFilterParamsCursor_EmptyIdsTreatedAsNoFilter(t *testing.T) {
	filter := &model.LeaderboardConfigFilter{
		Ids: []string{},
	}

	params, err := buildLeaderboardConfigFilterParamsCursor(filter, nil, nil, nil, nil)

	require.NoError(t, err)
	assert.Nil(t, params.Ids, "an explicitly empty ids slice should not filter out every row")
}

func TestBuildCountLeaderboardConfigsFilterParams_NilFilter(t *testing.T) {
	params := buildCountLeaderboardConfigsFilterParams(nil)
	assert.Equal(t, "", params.Projectid)
	assert.Nil(t, params.Isactive)
}

func TestBuildCountLeaderboardConfigsFilterParams_WithFilter(t *testing.T) {
	eventID := "EV01ARZ3NDEKTSV4RRFFQ69G5FAV"
	filter := &model.LeaderboardConfigFilter{
		EventID: &eventID,
		Ids:     []string{"LC1", "LC2"},
	}

	params := buildCountLeaderboardConfigsFilterParams(filter)

	assert.Equal(t, eventID, params.Eventid)
	assert.Equal(t, []string{"LC1", "LC2"}, params.Ids)
}

func TestBuildCountLeaderboardConfigsFilterParams_EmptyIdsTreatedAsNoFilter(t *testing.T) {
	filter := &model.LeaderboardConfigFilter{
		Ids: []string{},
	}

	params := buildCountLeaderboardConfigsFilterParams(filter)

	assert.Nil(t, params.Ids, "an explicitly empty ids slice should not filter out every row")
}
