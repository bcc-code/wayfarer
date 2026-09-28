package api

import (
	"testing"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPercentageOf(t *testing.T) {
	tests := []struct {
		name     string
		part     int
		total    int
		expected float64
	}{
		{"half", 5, 10, 50},
		{"all", 10, 10, 100},
		{"none", 0, 10, 0},
		{"rounds to one decimal", 1, 3, 33.3},
		{"zero total does not divide by zero", 3, 0, 0},
		{"negative total is treated as empty", 3, -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, percentageOf(tt.part, tt.total))
		})
	}
}

func TestNumberStats(t *testing.T) {
	t.Run("returns nil for every stat when there are no values", func(t *testing.T) {
		avg, median, minimum, maximum := numberStats(nil)
		assert.Nil(t, avg)
		assert.Nil(t, median)
		assert.Nil(t, minimum)
		assert.Nil(t, maximum)
	})

	t.Run("a single value is its own average, median and bounds", func(t *testing.T) {
		avg, median, minimum, maximum := numberStats([]float64{7})
		require.NotNil(t, avg)
		assert.Equal(t, 7.0, *avg)
		assert.Equal(t, 7.0, *median)
		assert.Equal(t, 7.0, *minimum)
		assert.Equal(t, 7.0, *maximum)
	})

	t.Run("odd count takes the middle value", func(t *testing.T) {
		avg, median, minimum, maximum := numberStats([]float64{5, 1, 3})
		assert.InDelta(t, 3.0, *avg, 0.0001)
		assert.Equal(t, 3.0, *median)
		assert.Equal(t, 1.0, *minimum)
		assert.Equal(t, 5.0, *maximum)
	})

	t.Run("even count averages the two middle values", func(t *testing.T) {
		_, median, _, _ := numberStats([]float64{1, 2, 3, 4})
		assert.Equal(t, 2.5, *median)
	})

	t.Run("does not mutate the caller's slice", func(t *testing.T) {
		values := []float64{5, 1, 3}
		numberStats(values)
		assert.Equal(t, []float64{5, 1, 3}, values)
	})
}

func TestNumberBuckets(t *testing.T) {
	t.Run("no values yields no buckets", func(t *testing.T) {
		assert.Empty(t, numberBuckets(nil, numberBucketTarget))
	})

	t.Run("identical values collapse to one full bucket", func(t *testing.T) {
		buckets := numberBuckets([]float64{40, 40, 40}, numberBucketTarget)
		require.Len(t, buckets, 1)
		assert.Equal(t, 40.0, buckets[0].From)
		assert.Equal(t, 40.0, buckets[0].To)
		assert.Equal(t, 3, buckets[0].Count)
		assert.Equal(t, 100.0, buckets[0].Percentage)
	})

	t.Run("every value lands in exactly one bucket", func(t *testing.T) {
		values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		buckets := numberBuckets(values, numberBucketTarget)

		total := 0
		for _, b := range buckets {
			total += b.Count
		}
		assert.Equal(t, len(values), total)
	})

	t.Run("the maximum is inside the top bucket, not past it", func(t *testing.T) {
		buckets := numberBuckets([]float64{0, 50, 100}, 3)
		require.NotEmpty(t, buckets)
		last := buckets[len(buckets)-1]
		assert.Equal(t, 100.0, last.To)
		assert.GreaterOrEqual(t, last.Count, 1)
	})

	t.Run("never makes more buckets than there are distinct values", func(t *testing.T) {
		buckets := numberBuckets([]float64{1, 1, 5, 5, 9}, numberBucketTarget)
		assert.LessOrEqual(t, len(buckets), 3)
	})

	t.Run("a target below one still produces a usable bucket", func(t *testing.T) {
		buckets := numberBuckets([]float64{1, 2, 3}, 0)
		require.Len(t, buckets, 1)
		assert.Equal(t, 3, buckets[0].Count)
	})
}

func TestNormalizeFreeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"case folded", "Håp", "håp"},
		{"trimmed", "  håp  ", "håp"},
		{"inner whitespace collapsed", "godt  fellesskap", "godt fellesskap"},
		{"tabs and newlines count as whitespace", "godt\tfellesskap\n", "godt fellesskap"},
		{"blank becomes empty", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeFreeText(tt.input))
		})
	}
}

func TestGroupFreeText(t *testing.T) {
	t.Run("no responses yields no groups", func(t *testing.T) {
		assert.Empty(t, groupFreeText(nil))
	})

	t.Run("groups answers that differ only in case and spacing", func(t *testing.T) {
		groups := groupFreeText([]string{"Håp", "håp", "  HÅP "})
		require.Len(t, groups, 1)
		assert.Equal(t, 3, groups[0].Count)
		assert.Equal(t, 100.0, groups[0].Percentage)
	})

	t.Run("labels a group with the spelling most people used", func(t *testing.T) {
		groups := groupFreeText([]string{"håp", "håp", "Håp"})
		require.Len(t, groups, 1)
		assert.Equal(t, "håp", groups[0].Text)
	})

	t.Run("orders by count descending", func(t *testing.T) {
		groups := groupFreeText([]string{"glede", "håp", "håp", "håp", "glede"})
		require.Len(t, groups, 2)
		assert.Equal(t, "håp", groups[0].Text)
		assert.Equal(t, 3, groups[0].Count)
		assert.Equal(t, "glede", groups[1].Text)
		assert.Equal(t, 2, groups[1].Count)
	})

	t.Run("breaks count ties alphabetically so output is stable", func(t *testing.T) {
		first := groupFreeText([]string{"beta", "alfa", "gamma"})
		second := groupFreeText([]string{"gamma", "beta", "alfa"})
		assert.Equal(t, first, second)
		require.Len(t, first, 3)
		assert.Equal(t, "alfa", first[0].Text)
		assert.Equal(t, "beta", first[1].Text)
		assert.Equal(t, "gamma", first[2].Text)
	})

	t.Run("percentages are a share of the grouped answers", func(t *testing.T) {
		groups := groupFreeText([]string{"a", "a", "a", "b"})
		require.Len(t, groups, 2)
		assert.Equal(t, 75.0, groups[0].Percentage)
		assert.Equal(t, 25.0, groups[1].Percentage)
	})

	t.Run("blank answers are dropped rather than grouped as empty", func(t *testing.T) {
		groups := groupFreeText([]string{"håp", "   ", ""})
		require.Len(t, groups, 1)
		assert.Equal(t, "håp", groups[0].Text)
		assert.Equal(t, 100.0, groups[0].Percentage)
	})
}

func TestOrderingAccuracy(t *testing.T) {
	correct := []string{"a", "b", "c"}

	t.Run("no submissions leaves every position at zero", func(t *testing.T) {
		assert.Equal(t, []int{0, 0, 0}, orderingAccuracy(nil, correct))
	})

	t.Run("a perfect answer counts for every position", func(t *testing.T) {
		got := orderingAccuracy([][]string{{"a", "b", "c"}}, correct)
		assert.Equal(t, []int{1, 1, 1}, got)
	})

	t.Run("counts only the positions actually placed right", func(t *testing.T) {
		// "a" right, "c" and "b" swapped.
		got := orderingAccuracy([][]string{{"a", "c", "b"}}, correct)
		assert.Equal(t, []int{1, 0, 0}, got)
	})

	t.Run("a short submission still counts for the positions it covers", func(t *testing.T) {
		got := orderingAccuracy([][]string{{"a", "b"}}, correct)
		assert.Equal(t, []int{1, 1, 0}, got)
	})

	t.Run("a longer submission does not read past the correct order", func(t *testing.T) {
		got := orderingAccuracy([][]string{{"a", "b", "c", "d"}}, correct)
		assert.Equal(t, []int{1, 1, 1}, got)
	})

	t.Run("aggregates across submissions", func(t *testing.T) {
		got := orderingAccuracy([][]string{
			{"a", "b", "c"},
			{"a", "c", "b"},
			{"b", "a", "c"},
		}, correct)
		assert.Equal(t, []int{2, 1, 2}, got)
	})

	t.Run("an empty correct order yields no positions", func(t *testing.T) {
		assert.Empty(t, orderingAccuracy([][]string{{"a"}}, nil))
	})
}

func TestDecodeOrderingSubmissions(t *testing.T) {
	t.Run("decodes arrays of ids", func(t *testing.T) {
		got := decodeOrderingSubmissions([][]byte{[]byte(`["a","b"]`)})
		assert.Equal(t, [][]string{{"a", "b"}}, got)
	})

	t.Run("skips a malformed row instead of failing the page", func(t *testing.T) {
		got := decodeOrderingSubmissions([][]byte{
			[]byte(`["a","b"]`),
			[]byte(`{"not":"an array"}`),
			[]byte(`nonsense`),
			[]byte(`["c"]`),
		})
		assert.Equal(t, [][]string{{"a", "b"}, {"c"}}, got)
	})
}

func TestBuildQuestionResults(t *testing.T) {
	const questionID = "QQ01ARZ3NDEKTSV4RRFFQ69G5FAV"

	countsFor := func(responses, correct int32) quizResultsInputs {
		return quizResultsInputs{
			responseCounts: map[string]*sqlc.GetQuizResponseCountsByQuestionRow{
				questionID: {QuestionID: questionID, ResponseCount: responses, CorrectCount: correct},
			},
			answerCounts: map[string]map[string]int{},
			numbers:      map[string][]float64{},
			freeText:     map[string][]string{},
			ordering:     map[string][][]byte{},
		}
	}

	t.Run("predefined keeps options nobody picked", func(t *testing.T) {
		in := countsFor(10, 7)
		in.answerCounts[questionID] = map[string]int{"QA1": 7, "QA2": 3}

		answers := []*model.QuizPredefinedAnswer{
			{ID: "QA1", AnswerText: "Oslo", AnswerOrder: 1, IsCorrectValue: true},
			{ID: "QA2", AnswerText: "Bergen", AnswerOrder: 2},
			{ID: "QA3", AnswerText: "Tromsø", AnswerOrder: 3},
		}

		got, ok := buildQuestionResults(&model.PredefinedQuestion{ID: questionID}, answers, in).(*model.PredefinedQuestionResults)
		require.True(t, ok)
		require.Len(t, got.Options, 3)
		assert.Equal(t, 7, got.CorrectCount)
		assert.Equal(t, 7, got.Options[0].Count)
		assert.Equal(t, 70.0, got.Options[0].Percentage)
		assert.True(t, got.Options[0].IsCorrect)
		assert.Equal(t, 0, got.Options[2].Count, "an unpicked option must still appear, at zero")
		assert.Equal(t, 0.0, got.Options[2].Percentage)
	})

	t.Run("number reports stats and buckets over its values", func(t *testing.T) {
		in := countsFor(4, 0)
		in.numbers[questionID] = []float64{10, 20, 30, 40}

		got, ok := buildQuestionResults(&model.NumberQuestion{ID: questionID}, nil, in).(*model.NumberQuestionResults)
		require.True(t, ok)
		assert.Equal(t, 4, got.ResponseCount)
		require.NotNil(t, got.Average)
		assert.Equal(t, 25.0, *got.Average)
		assert.Equal(t, 10.0, *got.Min)
		assert.Equal(t, 40.0, *got.Max)
		assert.NotEmpty(t, got.Buckets)
	})

	t.Run("number with no answers has null stats and no buckets", func(t *testing.T) {
		got, ok := buildQuestionResults(&model.NumberQuestion{ID: questionID}, nil, countsFor(0, 0)).(*model.NumberQuestionResults)
		require.True(t, ok)
		assert.Zero(t, got.ResponseCount)
		assert.Nil(t, got.Average)
		assert.Empty(t, got.Buckets)
	})

	t.Run("free text groups and keeps the raw answers", func(t *testing.T) {
		in := countsFor(3, 0)
		in.freeText[questionID] = []string{"Håp", "håp", "glede"}

		got, ok := buildQuestionResults(&model.FreeTextQuestion{ID: questionID}, nil, in).(*model.FreeTextQuestionResults)
		require.True(t, ok)
		assert.Equal(t, 2, got.DistinctCount)
		assert.Len(t, got.Responses, 3, "the raw list keeps every answer as submitted")
		assert.Equal(t, 2, got.Groups[0].Count)
	})

	t.Run("ordering reports per-position accuracy against the correct order", func(t *testing.T) {
		in := countsFor(2, 1)
		in.ordering[questionID] = [][]byte{
			[]byte(`["QA1","QA2","QA3"]`),
			[]byte(`["QA1","QA3","QA2"]`),
		}

		items := []*model.QuizPredefinedAnswer{
			{ID: "QA1", AnswerText: "Skapelsen", AnswerOrder: 1},
			{ID: "QA2", AnswerText: "Utgangen", AnswerOrder: 2},
			{ID: "QA3", AnswerText: "Eksilet", AnswerOrder: 3},
		}

		got, ok := buildQuestionResults(&model.OrderingQuestion{ID: questionID}, items, in).(*model.OrderingQuestionResults)
		require.True(t, ok)
		assert.Equal(t, 1, got.FullyCorrectCount)
		require.Len(t, got.Items, 3)
		assert.Equal(t, 2, got.Items[0].CorrectlyPlacedCount)
		assert.Equal(t, 100.0, got.Items[0].Percentage)
		assert.Equal(t, 1, got.Items[1].CorrectlyPlacedCount)
		assert.Equal(t, 50.0, got.Items[1].Percentage)
		assert.Equal(t, 1, got.Items[0].CorrectPosition)
		assert.Equal(t, "Skapelsen", got.Items[0].Item.ItemText)
	})

	t.Run("json reports only how many answered", func(t *testing.T) {
		got, ok := buildQuestionResults(&model.JSONQuestion{ID: questionID}, nil, countsFor(8, 0)).(*model.JSONQuestionResults)
		require.True(t, ok)
		assert.Equal(t, 8, got.ResponseCount)
	})

	t.Run("a question nobody answered reports zero, not a missing entry", func(t *testing.T) {
		empty := quizResultsInputs{
			responseCounts: map[string]*sqlc.GetQuizResponseCountsByQuestionRow{},
			answerCounts:   map[string]map[string]int{},
			numbers:        map[string][]float64{},
			freeText:       map[string][]string{},
			ordering:       map[string][][]byte{},
		}

		got, ok := buildQuestionResults(&model.PredefinedQuestion{ID: questionID}, nil, empty).(*model.PredefinedQuestionResults)
		require.True(t, ok)
		assert.Zero(t, got.ResponseCount)
		assert.Zero(t, got.CorrectCount)
		assert.Empty(t, got.Options)
	})
}
