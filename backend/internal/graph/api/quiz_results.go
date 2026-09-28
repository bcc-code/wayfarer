package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// Shaping helpers for aggregated quiz results.
//
// Counting happens in SQL (queries/quiz_results.sql); everything here turns raw
// rows into what the admin panel draws. They are deliberately pure so the
// awkward cases — no responses, a single response, every answer identical — can
// be tested without a database.

// numberBucketTarget is how many histogram buckets a NUMBER question aims for.
// Few enough to read off a projector, enough to show a shape.
const numberBucketTarget = 6

// percentageOf returns part as a percentage of total, rounded to one decimal.
// A zero total yields 0 rather than NaN: a question nobody answered draws as an
// empty bar, not as a broken one.
func percentageOf(part, total int) float64 {
	if total <= 0 {
		return 0
	}
	return math.Round(float64(part)/float64(total)*1000) / 10
}

// numberStats returns the average, median, min and max of values.
// Returns nil for each when values is empty. Does not mutate values.
func numberStats(values []float64) (avg, median, minimum, maximum *float64) {
	if len(values) == 0 {
		return nil, nil, nil, nil
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	sum := 0.0
	for _, v := range sorted {
		sum += v
	}

	a := sum / float64(len(sorted))
	lo := sorted[0]
	hi := sorted[len(sorted)-1]

	var m float64
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		m = (sorted[mid-1] + sorted[mid]) / 2
	} else {
		m = sorted[mid]
	}

	return &a, &m, &lo, &hi
}

// numberBuckets splits values into up to `target` equal-width buckets spanning
// min..max. The top bucket is closed on both ends so the maximum value lands
// inside the histogram rather than falling off it.
//
// Values that are all identical produce a single bucket: an equal-width split of
// a zero-width range would divide by zero, and "everyone said 40" is a real and
// fairly common answer.
func numberBuckets(values []float64, target int) []model.NumberBucket {
	if len(values) == 0 {
		return []model.NumberBucket{}
	}
	if target < 1 {
		target = 1
	}

	lo, hi := values[0], values[0]
	for _, v := range values {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}

	if lo == hi {
		return []model.NumberBucket{{
			From:       lo,
			To:         hi,
			Count:      len(values),
			Percentage: 100,
		}}
	}

	// Never make more buckets than there are distinct values to put in them —
	// an eight-bucket histogram of three answers is mostly empty columns.
	if distinct := countDistinct(values); distinct < target {
		target = distinct
	}

	width := (hi - lo) / float64(target)
	buckets := make([]model.NumberBucket, target)
	for i := range buckets {
		buckets[i] = model.NumberBucket{
			From: lo + width*float64(i),
			To:   lo + width*float64(i+1),
		}
	}
	// Guard against float drift leaving the maximum just outside the last edge.
	buckets[target-1].To = hi

	for _, v := range values {
		idx := int((v - lo) / width)
		if idx >= target {
			idx = target - 1
		}
		if idx < 0 {
			idx = 0
		}
		buckets[idx].Count++
	}

	for i := range buckets {
		buckets[i].Percentage = percentageOf(buckets[i].Count, len(values))
	}

	return buckets
}

func countDistinct(values []float64) int {
	seen := make(map[float64]struct{}, len(values))
	for _, v := range values {
		seen[v] = struct{}{}
	}
	return len(seen)
}

// normalizeFreeText is the key free-text answers are grouped by: trimmed, inner
// whitespace collapsed, case-folded. Deliberately naive — exact match after
// normalisation, no stemming or fuzzy clustering — so a group always means
// "these people wrote the same thing".
func normalizeFreeText(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

// groupFreeText groups answers by their normalised form, most common first.
// The label shown for a group is the spelling most people actually used; ties
// there, and ties on count, break alphabetically so the output is stable across
// requests rather than following map iteration order.
func groupFreeText(responses []string) []model.FreeTextGroup {
	type group struct {
		total    int
		variants map[string]int
	}

	groups := make(map[string]*group)
	for _, raw := range responses {
		key := normalizeFreeText(raw)
		if key == "" {
			continue
		}
		g, ok := groups[key]
		if !ok {
			g = &group{variants: make(map[string]int)}
			groups[key] = g
		}
		g.total++
		g.variants[strings.TrimSpace(raw)]++
	}

	counted := 0
	for _, g := range groups {
		counted += g.total
	}

	out := make([]model.FreeTextGroup, 0, len(groups))
	for _, g := range groups {
		out = append(out, model.FreeTextGroup{
			Text:       dominantVariant(g.variants),
			Count:      g.total,
			Percentage: percentageOf(g.total, counted),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Text < out[j].Text
	})

	return out
}

// dominantVariant picks the most-used original spelling, alphabetical on a tie.
func dominantVariant(variants map[string]int) string {
	best := ""
	bestCount := -1
	for text, count := range variants {
		if count > bestCount || (count == bestCount && text < best) {
			best, bestCount = text, count
		}
	}
	return best
}

// orderingAccuracy reports, for each position in the correct order, how many
// submissions placed the right item there. `correct` is the item IDs in their
// correct order; each entry of `submitted` is one submission's order.
//
// Submissions of a different length than the correct order still count for the
// positions they do cover — a truncated answer is wrong, not unreadable.
func orderingAccuracy(submitted [][]string, correct []string) []int {
	perPosition := make([]int, len(correct))
	for _, order := range submitted {
		for i, id := range correct {
			if i < len(order) && order[i] == id {
				perPosition[i]++
			}
		}
	}
	return perPosition
}

// decodeOrderingSubmissions parses the stored JSON arrays of item IDs, skipping
// any row whose payload is not an array of strings. A malformed row is one
// person's answer, not a reason to fail the whole results page.
func decodeOrderingSubmissions(payloads [][]byte) [][]string {
	out := make([][]string, 0, len(payloads))
	for _, payload := range payloads {
		var order []string
		if err := json.Unmarshal(payload, &order); err != nil {
			continue
		}
		out = append(out, order)
	}
	return out
}

// quizResultsInputs is everything the six aggregate queries returned, keyed for
// per-question lookup.
type quizResultsInputs struct {
	responseCounts map[string]*sqlc.GetQuizResponseCountsByQuestionRow
	answerCounts   map[string]map[string]int
	numbers        map[string][]float64
	freeText       map[string][]string
	ordering       map[string][][]byte
}

// buildQuestionResults turns one question plus its aggregates into the result
// type matching its kind. `answers` carries predefined answers for PREDEFINED
// and ordering items for ORDERING — both live in quiz_predefined_answers, where
// answer_order doubles as the correct position.
func buildQuestionResults(
	question model.QuizQuestion,
	answers []*model.QuizPredefinedAnswer,
	in quizResultsInputs,
) model.QuizQuestionResults {
	id := question.GetID()

	responseCount := 0
	correctCount := 0
	if counts, ok := in.responseCounts[id]; ok {
		responseCount = int(counts.ResponseCount)
		correctCount = int(counts.CorrectCount)
	}

	switch question.(type) {
	case *model.PredefinedQuestion:
		return buildPredefinedResults(question, answers, in.answerCounts[id], responseCount, correctCount)
	case *model.NumberQuestion:
		values := in.numbers[id]
		avg, median, minimum, maximum := numberStats(values)
		return &model.NumberQuestionResults{
			Question:      question,
			ResponseCount: responseCount,
			Average:       avg,
			Median:        median,
			Min:           minimum,
			Max:           maximum,
			Buckets:       numberBuckets(values, numberBucketTarget),
		}
	case *model.FreeTextQuestion:
		responses := in.freeText[id]
		groups := groupFreeText(responses)
		return &model.FreeTextQuestionResults{
			Question:      question,
			ResponseCount: responseCount,
			DistinctCount: len(groups),
			Groups:        groups,
			Responses:     responses,
		}
	case *model.OrderingQuestion:
		return buildOrderingResults(question, answers, in.ordering[id], responseCount, correctCount)
	default:
		// JSON, and anything added to the enum before this switch catches up.
		return &model.JSONQuestionResults{
			Question:      question,
			ResponseCount: responseCount,
		}
	}
}

// buildPredefinedResults keeps every option the question offers, including ones
// nobody picked — an absent row means zero, not "leave it off the chart".
func buildPredefinedResults(
	question model.QuizQuestion,
	answers []*model.QuizPredefinedAnswer,
	counts map[string]int,
	responseCount, correctCount int,
) *model.PredefinedQuestionResults {
	options := make([]model.PredefinedOptionResult, len(answers))
	for i, answer := range answers {
		count := counts[answer.ID]
		options[i] = model.PredefinedOptionResult{
			Answer:     answer,
			Count:      count,
			Percentage: percentageOf(count, responseCount),
			IsCorrect:  answer.IsCorrectValue,
		}
	}

	return &model.PredefinedQuestionResults{
		Question:      question,
		ResponseCount: responseCount,
		CorrectCount:  correctCount,
		Options:       options,
	}
}

func buildOrderingResults(
	question model.QuizQuestion,
	items []*model.QuizPredefinedAnswer,
	payloads [][]byte,
	responseCount, correctCount int,
) *model.OrderingQuestionResults {
	// The loader returns items sorted by answer_order, which is the correct
	// sequence; index i is therefore position i.
	correct := make([]string, len(items))
	for i, item := range items {
		correct[i] = item.ID
	}

	perPosition := orderingAccuracy(decodeOrderingSubmissions(payloads), correct)

	results := make([]model.OrderingItemResult, len(items))
	for i, item := range items {
		results[i] = model.OrderingItemResult{
			Item: &model.QuizOrderingItem{
				ID:                item.ID,
				QuestionID:        item.QuestionID,
				ItemText:          item.AnswerText,
				CorrectOrderValue: item.AnswerOrder,
			},
			CorrectPosition:      item.AnswerOrder,
			CorrectlyPlacedCount: perPosition[i],
			Percentage:           percentageOf(perPosition[i], responseCount),
		}
	}

	return &model.OrderingQuestionResults{
		Question:          question,
		ResponseCount:     responseCount,
		FullyCorrectCount: correctCount,
		Items:             results,
	}
}

// loadQuizResultsInputs runs the five per-response aggregate queries and keys
// each result by question ID. Five queries for the whole page, whatever the
// question count — there is no per-question round trip here.
func (r *Resolver) loadQuizResultsInputs(ctx context.Context, quizID string) (quizResultsInputs, error) {
	in := quizResultsInputs{
		responseCounts: map[string]*sqlc.GetQuizResponseCountsByQuestionRow{},
		answerCounts:   map[string]map[string]int{},
		numbers:        map[string][]float64{},
		freeText:       map[string][]string{},
		ordering:       map[string][][]byte{},
	}

	counts, err := r.DB.Queries.GetQuizResponseCountsByQuestion(ctx, quizID)
	if err != nil {
		return in, fmt.Errorf("failed to load quiz response counts: %w", err)
	}
	for _, row := range counts {
		in.responseCounts[row.QuestionID] = row
	}

	answerCounts, err := r.DB.Queries.GetQuizPredefinedAnswerCounts(ctx, quizID)
	if err != nil {
		return in, fmt.Errorf("failed to load quiz answer counts: %w", err)
	}
	for _, row := range answerCounts {
		byAnswer, ok := in.answerCounts[row.QuestionID]
		if !ok {
			byAnswer = map[string]int{}
			in.answerCounts[row.QuestionID] = byAnswer
		}
		byAnswer[row.AnswerID] = int(row.Count)
	}

	numbers, err := r.DB.Queries.GetQuizNumberResponses(ctx, quizID)
	if err != nil {
		return in, fmt.Errorf("failed to load quiz number responses: %w", err)
	}
	for _, row := range numbers {
		in.numbers[row.QuestionID] = append(in.numbers[row.QuestionID], row.Value)
	}

	freeText, err := r.DB.Queries.GetQuizFreeTextResponses(ctx, quizID)
	if err != nil {
		return in, fmt.Errorf("failed to load quiz free text responses: %w", err)
	}
	for _, row := range freeText {
		in.freeText[row.QuestionID] = append(in.freeText[row.QuestionID], row.TextResponse)
	}

	ordering, err := r.DB.Queries.GetQuizOrderingResponses(ctx, quizID)
	if err != nil {
		return in, fmt.Errorf("failed to load quiz ordering responses: %w", err)
	}
	for _, row := range ordering {
		in.ordering[row.QuestionID] = append(in.ordering[row.QuestionID], row.SubmittedOrder)
	}

	return in, nil
}
