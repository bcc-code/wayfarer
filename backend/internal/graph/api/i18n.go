package api

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/middleware"
)

// i18n wrapper functions for loading entities with translations applied
// These functions load the base entity and merge in translations for the requested language

// applyStringTranslation returns the translated value if non-nil and non-empty, otherwise the base value
func applyStringTranslation(translated *string, base string) string {
	if translated != nil && *translated != "" {
		return *translated
	}
	return base
}

// applyHTMLTranslation returns the translated value as HTML if non-nil and non-empty, otherwise the base value
func applyHTMLTranslation(translated *string, base scalars.HTML) scalars.HTML {
	if translated != nil && *translated != "" {
		return scalars.HTML(*translated)
	}
	return base
}

// LoadProjectWithTranslation loads a project and applies translation for the requested language
func (r *Resolver) LoadProjectWithTranslation(ctx context.Context, id string) (*model.Project, error) {
	project, err := r.Loaders.ProjectByIDLoader.Load(ctx, id)()
	if err != nil {
		return nil, err
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return project, nil
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "project",
		EntityID:   id,
		LangCode:   lang,
	})()

	if trans == nil {
		return project, nil
	}

	// Create copy with translations applied
	translated := *project
	translated.Name = applyStringTranslation(trans.Name, project.Name)
	translated.Description = applyStringTranslation(trans.Description, project.Description)
	return &translated, nil
}

// LoadEventWithTranslation loads an event and applies translation for the requested language
func (r *Resolver) LoadEventWithTranslation(ctx context.Context, id string) (*model.Event, error) {
	event, err := r.Loaders.EventByIDLoader.Load(ctx, id)()
	if err != nil {
		return nil, err
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return event, nil
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "event",
		EntityID:   id,
		LangCode:   lang,
	})()

	if trans == nil {
		return event, nil
	}

	translated := *event
	translated.Name = applyStringTranslation(trans.Name, event.Name)
	translated.Description = applyStringTranslation(trans.Description, event.Description)
	return &translated, nil
}

// LoadChallengeWithTranslation loads a challenge and applies translation for the requested language
func (r *Resolver) LoadChallengeWithTranslation(ctx context.Context, id string) (model.Challenge, error) {
	challenge, err := r.Loaders.ChallengeByIDLoader.Load(ctx, id)()
	if err != nil {
		return nil, err
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return challenge, nil
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "challenge",
		EntityID:   id,
		LangCode:   lang,
	})()

	if trans == nil {
		return challenge, nil
	}

	return applyChallengeTranslation(challenge, trans), nil
}

// applyChallengeTranslation applies translation to a Challenge interface value
func applyChallengeTranslation(challenge model.Challenge, trans *loaders.Translation) model.Challenge {
	switch c := challenge.(type) {
	case *model.SimpleChallenge:
		translated := *c
		translated.Name = applyStringTranslation(trans.Name, c.Name)
		translated.Description = applyHTMLTranslation(trans.Description, c.Description)
		translated.ButtonText = applyStringTranslation(trans.ButtonText, c.ButtonText)
		translated.NotificationText = applyStringTranslation(trans.NotificationText, c.NotificationText)
		return &translated
	case *model.QuizChallenge:
		translated := *c
		translated.Name = applyStringTranslation(trans.Name, c.Name)
		translated.Description = applyHTMLTranslation(trans.Description, c.Description)
		translated.ButtonText = applyStringTranslation(trans.ButtonText, c.ButtonText)
		translated.NotificationText = applyStringTranslation(trans.NotificationText, c.NotificationText)
		return &translated
	case *model.ExternalChallenge:
		translated := *c
		translated.Name = applyStringTranslation(trans.Name, c.Name)
		translated.Description = applyHTMLTranslation(trans.Description, c.Description)
		translated.ButtonText = applyStringTranslation(trans.ButtonText, c.ButtonText)
		translated.NotificationText = applyStringTranslation(trans.NotificationText, c.NotificationText)
		return &translated
	case *model.PluginChallenge:
		translated := *c
		translated.Name = applyStringTranslation(trans.Name, c.Name)
		translated.Description = applyHTMLTranslation(trans.Description, c.Description)
		if c.ButtonText != nil {
			bt := applyStringTranslation(trans.ButtonText, *c.ButtonText)
			translated.ButtonText = &bt
		}
		translated.NotificationText = applyStringTranslation(trans.NotificationText, c.NotificationText)
		return &translated
	default:
		return challenge
	}
}

// ApplyTranslationToProject applies translation to an already-loaded project
func (r *Resolver) ApplyTranslationToProject(ctx context.Context, project *model.Project) *model.Project {
	if project == nil {
		return nil
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return project
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "project",
		EntityID:   project.ID,
		LangCode:   lang,
	})()

	if trans == nil {
		return project
	}

	translated := *project
	translated.Name = applyStringTranslation(trans.Name, project.Name)
	translated.Description = applyStringTranslation(trans.Description, project.Description)
	return &translated
}

// ApplyTranslationToEvent applies translation to an already-loaded event
func (r *Resolver) ApplyTranslationToEvent(ctx context.Context, event *model.Event) *model.Event {
	if event == nil {
		return nil
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return event
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "event",
		EntityID:   event.ID,
		LangCode:   lang,
	})()

	if trans == nil {
		return event
	}

	translated := *event
	translated.Name = applyStringTranslation(trans.Name, event.Name)
	translated.Description = applyStringTranslation(trans.Description, event.Description)
	return &translated
}

// ApplyTranslationToChallenge applies translation to an already-loaded challenge
func (r *Resolver) ApplyTranslationToChallenge(ctx context.Context, challenge model.Challenge) model.Challenge {
	if challenge == nil {
		return nil
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return challenge
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "challenge",
		EntityID:   getChallengeID(challenge),
		LangCode:   lang,
	})()

	if trans == nil {
		return challenge
	}

	return applyChallengeTranslation(challenge, trans)
}

// Note: Article and Track translations are now handled via ExternalContent
// The title field on Article and Track is resolved from ExternalContent translations

// LoadQuizWithTranslation loads a quiz and applies translation for the requested language
func (r *Resolver) LoadQuizWithTranslation(ctx context.Context, id string) (*model.Quiz, error) {
	quiz, err := r.Loaders.QuizByIDLoader.Load(ctx, id)()
	if err != nil {
		return nil, err
	}

	return r.ApplyTranslationToQuiz(ctx, quiz), nil
}

// ApplyTranslationToQuiz applies translation to an already-loaded quiz
func (r *Resolver) ApplyTranslationToQuiz(ctx context.Context, quiz *model.Quiz) *model.Quiz {
	if quiz == nil {
		return nil
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return quiz
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "quiz",
		EntityID:   quiz.ID,
		LangCode:   lang,
	})()

	if trans == nil {
		return quiz
	}

	translated := *quiz
	translated.Name = applyStringTranslation(trans.Name, quiz.Name)
	translated.Description = applyStringTranslation(trans.Description, quiz.Description)
	return &translated
}

// ApplyTranslationToQuizQuestion applies translation to an already-loaded quiz question
func (r *Resolver) ApplyTranslationToQuizQuestion(ctx context.Context, question model.QuizQuestion) model.QuizQuestion {
	if question == nil {
		return nil
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return question
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "quiz_question",
		EntityID:   question.GetID(),
		LangCode:   lang,
	})()

	if trans == nil {
		return question
	}

	return applyQuizQuestionTranslation(question, trans)
}

// applyQuizQuestionTranslation applies translation to a QuizQuestion interface value
func applyQuizQuestionTranslation(question model.QuizQuestion, trans *loaders.Translation) model.QuizQuestion {
	switch q := question.(type) {
	case *model.FreeTextQuestion:
		translated := *q
		translated.QuestionText = applyStringTranslation(trans.QuestionText, q.QuestionText)
		return &translated
	case *model.PredefinedQuestion:
		translated := *q
		translated.QuestionText = applyStringTranslation(trans.QuestionText, q.QuestionText)
		return &translated
	case *model.NumberQuestion:
		translated := *q
		translated.QuestionText = applyStringTranslation(trans.QuestionText, q.QuestionText)
		return &translated
	case *model.JSONQuestion:
		translated := *q
		translated.QuestionText = applyStringTranslation(trans.QuestionText, q.QuestionText)
		return &translated
	case *model.OrderingQuestion:
		translated := *q
		translated.QuestionText = applyStringTranslation(trans.QuestionText, q.QuestionText)
		return &translated
	default:
		return question
	}
}

// ApplyTranslationToQuizAnswer applies translation to an already-loaded quiz predefined answer
func (r *Resolver) ApplyTranslationToQuizAnswer(ctx context.Context, answer *model.QuizPredefinedAnswer) *model.QuizPredefinedAnswer {
	if answer == nil {
		return nil
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return answer
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "quiz_answer",
		EntityID:   answer.ID,
		LangCode:   lang,
	})()

	if trans == nil {
		return answer
	}

	translated := *answer
	translated.AnswerText = applyStringTranslation(trans.AnswerText, answer.AnswerText)
	return &translated
}

// LoadConsentWithTranslation loads a consent and applies translation for the requested language
func (r *Resolver) LoadConsentWithTranslation(ctx context.Context, id string) (*model.Consent, error) {
	consent, err := r.Loaders.ConsentByIDLoader.Load(ctx, id)()
	if err != nil {
		return nil, err
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return consent, nil
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "consent",
		EntityID:   id,
		LangCode:   lang,
	})()

	if trans == nil {
		return consent, nil
	}

	// Create copy with translations applied
	translated := *consent
	translated.Title = applyStringTranslation(trans.Title, consent.Title)
	translated.ShortText = applyStringTranslation(trans.ShortText, consent.ShortText)
	translated.BodyMarkdown = applyStringTranslation(trans.Body, consent.BodyMarkdown)
	return &translated, nil
}

// ApplyTranslationToConsent applies translation to an already-loaded consent
func (r *Resolver) ApplyTranslationToConsent(ctx context.Context, consent *model.Consent) *model.Consent {
	if consent == nil {
		return nil
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return consent
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "consent",
		EntityID:   consent.ID,
		LangCode:   lang,
	})()

	if trans == nil {
		return consent
	}

	translated := *consent
	translated.Title = applyStringTranslation(trans.Title, consent.Title)
	translated.ShortText = applyStringTranslation(trans.ShortText, consent.ShortText)
	translated.BodyMarkdown = applyStringTranslation(trans.Body, consent.BodyMarkdown)
	return &translated
}

// LoadAchievementWithTranslation loads an achievement and applies translation for the requested language
func (r *Resolver) LoadAchievementWithTranslation(ctx context.Context, id string) (model.Achievement, error) {
	achievementThunk := r.Loaders.AchievementByIDLoader.Load(ctx, id)
	achievement, err := achievementThunk()
	if err != nil {
		return nil, err
	}

	return r.ApplyTranslationToAchievement(ctx, achievement), nil
}

// ApplyTranslationToAchievement applies translation to an already-loaded achievement
func (r *Resolver) ApplyTranslationToAchievement(ctx context.Context, achievement model.Achievement) model.Achievement {
	if achievement == nil {
		return nil
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return achievement
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "achievement",
		EntityID:   getAchievementID(achievement),
		LangCode:   lang,
	})()

	if trans == nil {
		return achievement
	}

	return applyAchievementTranslation(achievement, trans)
}

// getAchievementID extracts the ID from any Achievement implementation
func getAchievementID(a model.Achievement) string {
	switch v := a.(type) {
	case *model.SimpleAchievement:
		return v.ID
	case *model.ContentAchievement:
		return v.ID
	case *model.StreakAchievement:
		return v.ID
	default:
		return ""
	}
}

// applyAchievementTranslation applies translation to an Achievement interface value
func applyAchievementTranslation(achievement model.Achievement, trans *loaders.Translation) model.Achievement {
	switch a := achievement.(type) {
	case *model.SimpleAchievement:
		translated := *a
		translated.Name = applyStringTranslation(trans.Name, a.Name)
		translated.DescriptionPending = applyStringTranslation(trans.DescriptionPending, a.DescriptionPending)
		translated.DescriptionCompleted = applyStringTranslation(trans.DescriptionCompleted, a.DescriptionCompleted)
		translated.NotificationText = applyStringTranslation(trans.NotificationText, a.NotificationText)
		return &translated
	case *model.ContentAchievement:
		translated := *a
		translated.Name = applyStringTranslation(trans.Name, a.Name)
		translated.DescriptionPending = applyStringTranslation(trans.DescriptionPending, a.DescriptionPending)
		translated.DescriptionCompleted = applyStringTranslation(trans.DescriptionCompleted, a.DescriptionCompleted)
		translated.NotificationText = applyStringTranslation(trans.NotificationText, a.NotificationText)
		return &translated
	case *model.StreakAchievement:
		translated := *a
		translated.Name = applyStringTranslation(trans.Name, a.Name)
		translated.DescriptionPending = applyStringTranslation(trans.DescriptionPending, a.DescriptionPending)
		translated.DescriptionCompleted = applyStringTranslation(trans.DescriptionCompleted, a.DescriptionCompleted)
		translated.NotificationText = applyStringTranslation(trans.NotificationText, a.NotificationText)
		return &translated
	default:
		return achievement
	}
}
