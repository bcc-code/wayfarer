package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuizReorderPasses(t *testing.T) {
	passes := quizReorderPasses([]string{"QQ3", "QQ1", "QQ2"})

	require.Len(t, passes, 2)

	// Every question is parked on an order no real position uses, so the
	// second pass cannot collide with a position still held by another row.
	assert.Equal(t, []quizReorderStep{
		{ID: "QQ3", Order: -1},
		{ID: "QQ1", Order: -2},
		{ID: "QQ2", Order: -3},
	}, passes[0])

	// 1-based, matching the orders the admin panel sends when adding a question.
	assert.Equal(t, []quizReorderStep{
		{ID: "QQ3", Order: 1},
		{ID: "QQ1", Order: 2},
		{ID: "QQ2", Order: 3},
	}, passes[1])
}

func TestQuizReorderPassesEmpty(t *testing.T) {
	assert.Nil(t, quizReorderPasses(nil))
	assert.Nil(t, quizReorderPasses([]string{}))
}

// A swap is the case the single-pass version got wrong.
func TestQuizReorderPassesSwapNeverReusesALiveOrder(t *testing.T) {
	passes := quizReorderPasses([]string{"QQ2", "QQ1"})

	parked := map[int32]bool{}
	for _, step := range passes[0] {
		assert.Negative(t, step.Order)
		parked[step.Order] = true
	}
	assert.Len(t, parked, 2, "parked orders must be distinct")

	for _, step := range passes[1] {
		assert.False(t, parked[step.Order])
	}
}
