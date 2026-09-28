package services

import (
	"fmt"
	"time"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
)

// IsProjectFinished checks if a project is past its end date or archived.
// Returns an error if the project is finished and should not allow new awards.
func IsProjectFinished(project *model.Project) error {
	if project == nil {
		return fmt.Errorf("project not found")
	}
	if project.ArchivedAt != nil && *project.ArchivedAt {
		return fmt.Errorf("project is archived and no longer accepting achievement awards")
	}
	if time.Now().After(project.EndDate.Time) {
		return fmt.Errorf("project has ended and is no longer accepting achievement awards")
	}
	return nil
}
