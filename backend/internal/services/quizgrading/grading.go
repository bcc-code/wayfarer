// Package quizgrading decides whether a quiz response is correct.
//
// Grading is all or nothing: a response is either fully correct or wrong.
// It is the single source of truth for is_correct, used by every resolver
// that stores a quiz response.
package quizgrading

// Question types that are graded automatically.
const (
	QuestionTypePredefined = "PREDEFINED"
	QuestionTypeOrdering   = "ORDERING"
)

// Answer is one predefined answer of a question.
// Answers must be in answer_order, which is the correct sequence for ORDERING.
type Answer struct {
	ID        string
	IsCorrect bool
}

// Response holds the gradable parts of a submitted answer.
type Response struct {
	SelectedAnswerIDs []string // PREDEFINED
	SubmittedOrder    []string // ORDERING
}

// IsGradable reports whether responses to the question type are graded,
// i.e. whether Grade needs the question's answers.
func IsGradable(questionType string) bool {
	return questionType == QuestionTypePredefined || questionType == QuestionTypeOrdering
}

// Grade returns whether the response is correct, or nil when it is not graded:
// FREE_TEXT, NUMBER and JSON questions, or a gradable question answered
// without the field it needs.
func Grade(questionType string, response Response, answers []Answer) *bool {
	var correct bool
	switch questionType {
	case QuestionTypePredefined:
		if response.SelectedAnswerIDs == nil {
			return nil
		}
		correct = SelectionCorrect(response.SelectedAnswerIDs, answers)
	case QuestionTypeOrdering:
		if response.SubmittedOrder == nil {
			return nil
		}
		correct = OrderCorrect(response.SubmittedOrder, answers)
	default:
		return nil
	}
	return &correct
}

// SelectionCorrect reports whether exactly the correct answers were selected:
// every correct answer and no wrong one. Works for single and multiple choice.
func SelectionCorrect(selectedIDs []string, answers []Answer) bool {
	correctIDs := make(map[string]bool)
	for _, a := range answers {
		if a.IsCorrect {
			correctIDs[a.ID] = true
		}
	}

	selected := make(map[string]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		if !correctIDs[id] {
			return false
		}
		selected[id] = true
	}
	return len(selected) == len(correctIDs)
}

// OrderCorrect reports whether every item is in its correct position.
func OrderCorrect(submittedOrder []string, answers []Answer) bool {
	if len(submittedOrder) != len(answers) {
		return false
	}
	for i, a := range answers {
		if submittedOrder[i] != a.ID {
			return false
		}
	}
	return true
}
