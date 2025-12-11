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

// applyMarkdownTranslation returns the translated value as Markdown if non-nil and non-empty, otherwise the base value
func applyMarkdownTranslation(translated *string, base scalars.Markdown) scalars.Markdown {
	if translated != nil && *translated != "" {
		return scalars.Markdown(*translated)
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

// LoadTeamWithTranslation loads a team and applies translation for the requested language
func (r *Resolver) LoadTeamWithTranslation(ctx context.Context, id string) (*model.Team, error) {
	team, err := r.Loaders.TeamByIDLoader.Load(ctx, id)()
	if err != nil {
		return nil, err
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return team, nil
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "team",
		EntityID:   id,
		LangCode:   lang,
	})()

	if trans == nil {
		return team, nil
	}

	translated := *team
	translated.Name = applyStringTranslation(trans.Name, team.Name)
	translated.Description = applyStringTranslation(trans.Description, team.Description)
	return &translated, nil
}

// LoadSuperTeamWithTranslation loads a super team and applies translation for the requested language
func (r *Resolver) LoadSuperTeamWithTranslation(ctx context.Context, id string) (*model.SuperTeam, error) {
	superTeam, err := r.Loaders.SuperTeamByIDLoader.Load(ctx, id)()
	if err != nil {
		return nil, err
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return superTeam, nil
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "superteam",
		EntityID:   id,
		LangCode:   lang,
	})()

	if trans == nil {
		return superTeam, nil
	}

	translated := *superTeam
	translated.Name = applyStringTranslation(trans.Name, superTeam.Name)
	translated.Description = applyStringTranslation(trans.Description, superTeam.Description)
	return &translated, nil
}

// LoadStreakWithTranslation loads a streak and applies translation for the requested language
func (r *Resolver) LoadStreakWithTranslation(ctx context.Context, id string) (*model.Streak, error) {
	streak, err := r.Loaders.StreakByIDLoader.Load(ctx, id)()
	if err != nil {
		return nil, err
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return streak, nil
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "streak",
		EntityID:   id,
		LangCode:   lang,
	})()

	if trans == nil {
		return streak, nil
	}

	translated := *streak
	translated.Name = applyStringTranslation(trans.Name, streak.Name)
	translated.Description = applyStringTranslation(trans.Description, streak.Description)
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
		return &translated
	case *model.QuizChallenge:
		translated := *c
		translated.Name = applyStringTranslation(trans.Name, c.Name)
		translated.Description = applyHTMLTranslation(trans.Description, c.Description)
		translated.ButtonText = applyStringTranslation(trans.ButtonText, c.ButtonText)
		return &translated
	case *model.ExternalChallenge:
		translated := *c
		translated.Name = applyStringTranslation(trans.Name, c.Name)
		translated.Description = applyHTMLTranslation(trans.Description, c.Description)
		translated.ButtonText = applyStringTranslation(trans.ButtonText, c.ButtonText)
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

// ApplyTranslationToTeam applies translation to an already-loaded team
func (r *Resolver) ApplyTranslationToTeam(ctx context.Context, team *model.Team) *model.Team {
	if team == nil {
		return nil
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return team
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "team",
		EntityID:   team.ID,
		LangCode:   lang,
	})()

	if trans == nil {
		return team
	}

	translated := *team
	translated.Name = applyStringTranslation(trans.Name, team.Name)
	translated.Description = applyStringTranslation(trans.Description, team.Description)
	return &translated
}

// ApplyTranslationToSuperTeam applies translation to an already-loaded super team
func (r *Resolver) ApplyTranslationToSuperTeam(ctx context.Context, superTeam *model.SuperTeam) *model.SuperTeam {
	if superTeam == nil {
		return nil
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return superTeam
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "superteam",
		EntityID:   superTeam.ID,
		LangCode:   lang,
	})()

	if trans == nil {
		return superTeam
	}

	translated := *superTeam
	translated.Name = applyStringTranslation(trans.Name, superTeam.Name)
	translated.Description = applyStringTranslation(trans.Description, superTeam.Description)
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

// ApplyTranslationToStreak applies translation to an already-loaded streak
func (r *Resolver) ApplyTranslationToStreak(ctx context.Context, streak *model.Streak) *model.Streak {
	if streak == nil {
		return nil
	}

	lang := middleware.GetLanguage(ctx)
	if lang == middleware.DefaultLanguage {
		return streak
	}

	trans, _ := r.Loaders.TranslationLoader.Load(ctx, loaders.TranslationKey{
		EntityType: "streak",
		EntityID:   streak.ID,
		LangCode:   lang,
	})()

	if trans == nil {
		return streak
	}

	translated := *streak
	translated.Name = applyStringTranslation(trans.Name, streak.Name)
	translated.Description = applyStringTranslation(trans.Description, streak.Description)
	return &translated
}

// Note: Article and Track translations are now handled via ExternalContent
// The title field on Article and Track is resolved from ExternalContent translations

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
