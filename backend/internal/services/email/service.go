package email

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/resend/resend-go/v2"
)

// Service handles email operations using Resend
type Service struct {
	client       *resend.Client
	adminBaseURL string
}

// NewService creates a new email service
func NewService(apiKey string, adminBaseURL string) *Service {
	if apiKey == "" {
		return &Service{
			client:       nil,
			adminBaseURL: adminBaseURL,
		}
	}
	return &Service{
		client:       resend.NewClient(apiKey),
		adminBaseURL: adminBaseURL,
	}
}

// ConsentEntry represents a consent action for email formatting
type ConsentEntry struct {
	Title  string
	Action string // "Accepted" or "Rejected"
	Date   time.Time
}

// FeedbackEmailParams contains all data needed to send a feedback email
type FeedbackEmailParams struct {
	UserID      string
	UserName    string
	UserEmail   string
	ChurchName  string
	TotalPoints int64
	ProjectName string // empty if no project context
	UserCreated time.Time
	Consents    []ConsentEntry
	Message     string
}

// SendFeedbackToDesk sends a feedback email to the support desk
func (s *Service) SendFeedbackToDesk(ctx context.Context, params FeedbackEmailParams) error {
	if s.client == nil {
		return fmt.Errorf("email service not configured: RESEND_API_KEY not set")
	}

	body := s.buildFeedbackEmailBody(params)

	sendParams := &resend.SendEmailRequest{
		From:    "Wayfarer Feedback <noreply@mailer.bcc.media>",
		To:      []string{"support@bcc.media"},
		ReplyTo: params.UserEmail,
		Subject: fmt.Sprintf("Interact feedback from %s", params.UserName),
		Text:    body,
	}

	_, err := s.client.Emails.Send(sendParams)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *Service) buildFeedbackEmailBody(params FeedbackEmailParams) string {
	var builder strings.Builder

	// Message
	builder.WriteString("Message:\n")
	builder.WriteString(params.Message)
	builder.WriteString("\n")
	builder.WriteString("-----------------------------------------------------\n")

	// User info with admin link
	userLink := params.UserID
	if s.adminBaseURL != "" {
		userLink = fmt.Sprintf("%s (%s/admin/users/%s)", params.UserID, s.adminBaseURL, params.UserID)
	}
	builder.WriteString(fmt.Sprintf("User: %s\n", userLink))
	builder.WriteString(fmt.Sprintf("Church: %s\n", params.ChurchName))

	// Points with project context
	if params.ProjectName != "" {
		builder.WriteString(fmt.Sprintf("Total Points: %d (in %s)\n", params.TotalPoints, params.ProjectName))
	} else {
		builder.WriteString("Total Points: N/A (no project context)\n")
	}

	builder.WriteString(fmt.Sprintf("Created: %s\n", params.UserCreated.Format("2006-01-02")))
	builder.WriteString("\n")

	// Consents
	builder.WriteString("Consents:\n")
	if len(params.Consents) == 0 {
		builder.WriteString("  No consent records\n")
	} else {
		for _, consent := range params.Consents {
			builder.WriteString(fmt.Sprintf("  - %s: %s on %s\n",
				consent.Title,
				consent.Action,
				consent.Date.Format("2006-01-02"),
			))
		}
	}
	builder.WriteString("\n")

	return builder.String()
}
