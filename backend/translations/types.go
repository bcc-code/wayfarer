package translations

import "gopkg.in/guregu/null.v4"

// NameDescriptionTranslation is used for entities with name and description fields
// Used by: Projects, Events, Teams, SuperTeams, Streaks
type NameDescriptionTranslation struct {
	Name        string      `json:"name"`
	Description null.String `json:"description"`
}

// ChallengeTranslation includes button_text in addition to name and description
type ChallengeTranslation struct {
	Name        string      `json:"name"`
	Description null.String `json:"description"`
	ButtonText  string      `json:"button_text"`
}

// AchievementTranslation includes multiple description variants
type AchievementTranslation struct {
	Name                 string      `json:"name"`
	DescriptionPending   string      `json:"description_pending"`
	DescriptionCompleted string      `json:"description_completed"`
	NotificationText     null.String `json:"notification_text"`
}

// ConsentTranslation includes title, short_text, and body
type ConsentTranslation struct {
	Title     string `json:"title"`
	ShortText string `json:"short_text"`
	Body      string `json:"body"`
}

// QuizTranslation includes nested questions and answers
type QuizTranslation struct {
	Name        string           `json:"name"`
	Description null.String      `json:"description"`
	Questions   []QuestionWithID `json:"questions,omitempty"`
}

// QuestionWithID represents a quiz question with its ID for tracking
type QuestionWithID struct {
	QuestionText string         `json:"question_text"`
	ID           string         `json:"@id"` // Special field for Phrase to track question ID
	Answers      []AnswerWithID `json:"answers,omitempty"`
}

// AnswerWithID represents a quiz answer with its ID for tracking
type AnswerWithID struct {
	AnswerText string `json:"answer_text"`
	ID         string `json:"@id"` // Special field for Phrase to track answer ID
}
