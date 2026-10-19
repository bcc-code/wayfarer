package translations

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestAnswerWithIDJSON(t *testing.T) {
	t.Run("answer without description exports without the key", func(t *testing.T) {
		got, err := json.Marshal(AnswerWithID{AnswerText: "Melkekrukken", ID: "QA1"})
		require.NoError(t, err)

		// Byte-identical to the export before descriptions existed, so the
		// collection hash doesn't change for quizzes without descriptions.
		assert.JSONEq(t, `{"answer_text":"Melkekrukken","@id":"QA1"}`, string(got))
		assert.NotContains(t, string(got), "description")
	})

	t.Run("answer with description exports it", func(t *testing.T) {
		got, err := json.Marshal(AnswerWithID{
			AnswerText:  "Melkekrukken",
			Description: null.StringFrom("En krukke til oppbevaring av melk."),
			ID:          "QA1",
		})
		require.NoError(t, err)

		assert.JSONEq(t, `{"answer_text":"Melkekrukken","description":"En krukke til oppbevaring av melk.","@id":"QA1"}`, string(got))
	})

	t.Run("imported answer without description has an empty description", func(t *testing.T) {
		var a AnswerWithID
		require.NoError(t, json.Unmarshal([]byte(`{"answer_text":"Milk jar","@id":"QA1"}`), &a))

		assert.Equal(t, "Milk jar", a.AnswerText)
		assert.Equal(t, "", a.Description.String)
	})

	t.Run("imported answer with description keeps it", func(t *testing.T) {
		var a AnswerWithID
		require.NoError(t, json.Unmarshal([]byte(`{"answer_text":"Milk jar","description":"A jar for storing milk.","@id":"QA1"}`), &a))

		assert.Equal(t, "A jar for storing milk.", a.Description.String)
	})
}
