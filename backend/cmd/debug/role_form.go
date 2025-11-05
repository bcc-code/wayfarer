package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/ulid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackc/pgx/v5/pgtype"
)

type roleFormModel struct {
	db             *database.DB
	fields         []string
	cursor         int
	roles          []string
	roleIdx        int
	users          []string
	allUsers       []string
	userIdx        int
	userFilter     string
	assigners      []string
	allAssigners   []string
	assignerIdx    int
	assignerFilter string
	churches       []string
	churchIdx      int
	projects       []string
	projectIdx     int
	teams          []string
	teamIdx        int
	loadingMsg     string
	scopeHelp      string
}

func newRoleForm(db *database.DB) *roleFormModel {
	form := &roleFormModel{
		db: db,
		// Fields in order: role, user, assigner, then scope fields
		fields:      []string{"role", "user", "assigned_by", "church", "project", "team"},
		cursor:      0,
		roles:       []string{"SUPERADMIN", "ADMIN", "CHURCH_ADMIN", "PROJECT_ADMIN", "TEAM_LEAD", "USER", "M2M"},
		roleIdx:     5, // Default to USER
		users:       []string{},
		userIdx:     0,
		assigners:   []string{},
		assignerIdx: 0,
		churches:    []string{"(none)"},
		churchIdx:   0,
		projects:    []string{"(none)"},
		projectIdx:  0,
		teams:       []string{"(none)"},
		teamIdx:     0,
	}

	form.updateScopeHelp()

	// Load data in background
	go form.loadData()

	return form
}

func (f *roleFormModel) loadData() {
	ctx := context.Background()

	// Load users
	userQuery := `SELECT id, name, email FROM users ORDER BY name`
	rows, err := f.db.Pool.Query(ctx, userQuery)
	if err != nil {
		f.loadingMsg = fmt.Sprintf("Error loading users: %v", err)
		return
	}
	users := []string{}
	for rows.Next() {
		var id, name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			continue
		}
		users = append(users, fmt.Sprintf("%s <%s> (%s)", name, email, id))
	}
	rows.Close()

	if len(users) == 0 {
		f.loadingMsg = "No users found. Please add users first."
		return
	}

	f.users = users
	f.allUsers = users
	f.assigners = users
	f.allAssigners = users

	// Load churches
	churchQuery := `SELECT id, name FROM churches ORDER BY name`
	rows, err = f.db.Pool.Query(ctx, churchQuery)
	if err == nil {
		churches := []string{"(none)"}
		for rows.Next() {
			var id, name string
			if err := rows.Scan(&id, &name); err != nil {
				continue
			}
			churches = append(churches, fmt.Sprintf("%s (%s)", name, id))
		}
		rows.Close()
		f.churches = churches
	}

	// Load projects
	projectQuery := `SELECT id, name FROM projects ORDER BY name`
	rows, err = f.db.Pool.Query(ctx, projectQuery)
	if err == nil {
		projects := []string{"(none)"}
		for rows.Next() {
			var id, name string
			if err := rows.Scan(&id, &name); err != nil {
				continue
			}
			projects = append(projects, fmt.Sprintf("%s (%s)", name, id))
		}
		rows.Close()
		f.projects = projects
	}

	// Load teams
	teamQuery := `SELECT id, name FROM teams ORDER BY name`
	rows, err = f.db.Pool.Query(ctx, teamQuery)
	if err == nil {
		teams := []string{"(none)"}
		for rows.Next() {
			var id, name string
			if err := rows.Scan(&id, &name); err != nil {
				continue
			}
			teams = append(teams, fmt.Sprintf("%s (%s)", name, id))
		}
		rows.Close()
		f.teams = teams
	}

	f.loadingMsg = ""
}

func (f *roleFormModel) updateScopeHelp() {
	role := f.roles[f.roleIdx]
	switch role {
	case "SUPERADMIN", "ADMIN", "USER", "M2M":
		f.scopeHelp = "Global role - no scope needed"
	case "CHURCH_ADMIN":
		f.scopeHelp = "Requires church scope only"
	case "PROJECT_ADMIN":
		f.scopeHelp = "Requires project scope only"
	case "TEAM_LEAD":
		f.scopeHelp = "Requires team scope only"
	default:
		f.scopeHelp = ""
	}
}

func (f *roleFormModel) getSelectedUserID() string {
	if len(f.users) == 0 {
		return ""
	}
	return extractID(f.users[f.userIdx])
}

func (f *roleFormModel) getSelectedAssignerID() string {
	if len(f.assigners) == 0 {
		return ""
	}
	return extractID(f.assigners[f.assignerIdx])
}

func (f *roleFormModel) getSelectedChurchID() *string {
	if f.churchIdx == 0 {
		return nil
	}
	id := extractID(f.churches[f.churchIdx])
	return &id
}

func (f *roleFormModel) getSelectedProjectID() *string {
	if f.projectIdx == 0 {
		return nil
	}
	id := extractID(f.projects[f.projectIdx])
	return &id
}

func (f *roleFormModel) getSelectedTeamID() *string {
	if f.teamIdx == 0 {
		return nil
	}
	id := extractID(f.teams[f.teamIdx])
	return &id
}

func extractID(s string) string {
	start := strings.LastIndex(s, "(")
	end := strings.LastIndex(s, ")")
	if start != -1 && end != -1 {
		return s[start+1 : end]
	}
	return ""
}

func (f *roleFormModel) applyUserFilter() {
	if f.userFilter == "" {
		f.users = f.allUsers
		return
	}

	filtered := []string{}
	filter := strings.ToLower(f.userFilter)
	for _, user := range f.allUsers {
		if strings.Contains(strings.ToLower(user), filter) {
			filtered = append(filtered, user)
		}
	}

	if len(filtered) > 0 {
		f.users = filtered
		if f.userIdx >= len(filtered) {
			f.userIdx = len(filtered) - 1
		}
	} else {
		f.users = f.allUsers
	}
}

func (f *roleFormModel) applyAssignerFilter() {
	if f.assignerFilter == "" {
		f.assigners = f.allAssigners
		return
	}

	filtered := []string{}
	filter := strings.ToLower(f.assignerFilter)
	for _, assigner := range f.allAssigners {
		if strings.Contains(strings.ToLower(assigner), filter) {
			filtered = append(filtered, assigner)
		}
	}

	if len(filtered) > 0 {
		f.assigners = filtered
		if f.assignerIdx >= len(filtered) {
			f.assignerIdx = len(filtered) - 1
		}
	} else {
		f.assigners = f.allAssigners
	}
}

func (m model) updateRoleForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := m.roleForm

	switch msg.String() {
	case "ctrl+c", "esc":
		m.screen = ScreenMenu
		m.cursor = 0
		return m, nil

	case "up":
		if form.cursor > 0 {
			form.cursor--
		}

	case "down":
		if form.cursor < len(form.fields) {
			form.cursor++
		}

	case "ctrl+d": // Clear filter
		field := form.fields[form.cursor]
		if field == "user" {
			form.userFilter = ""
			form.applyUserFilter()
		} else if field == "assigned_by" {
			form.assignerFilter = ""
			form.applyAssignerFilter()
		}

	case "backspace":
		field := form.fields[form.cursor]
		if field == "user" && len(form.userFilter) > 0 {
			form.userFilter = form.userFilter[:len(form.userFilter)-1]
			form.applyUserFilter()
		} else if field == "assigned_by" && len(form.assignerFilter) > 0 {
			form.assignerFilter = form.assignerFilter[:len(form.assignerFilter)-1]
			form.applyAssignerFilter()
		}

	case "left":
		field := form.fields[form.cursor]
		switch field {
		case "role":
			if form.roleIdx > 0 {
				form.roleIdx--
				form.updateScopeHelp()
			}
		case "user":
			if form.userIdx > 0 {
				form.userIdx--
			}
		case "assigned_by":
			if form.assignerIdx > 0 {
				form.assignerIdx--
			}
		case "church":
			if form.churchIdx > 0 {
				form.churchIdx--
			}
		case "project":
			if form.projectIdx > 0 {
				form.projectIdx--
			}
		case "team":
			if form.teamIdx > 0 {
				form.teamIdx--
			}
		}

	case "right":
		field := form.fields[form.cursor]
		switch field {
		case "role":
			if form.roleIdx < len(form.roles)-1 {
				form.roleIdx++
				form.updateScopeHelp()
			}
		case "user":
			if form.userIdx < len(form.users)-1 {
				form.userIdx++
			}
		case "assigned_by":
			if form.assignerIdx < len(form.assigners)-1 {
				form.assignerIdx++
			}
		case "church":
			if form.churchIdx < len(form.churches)-1 {
				form.churchIdx++
			}
		case "project":
			if form.projectIdx < len(form.projects)-1 {
				form.projectIdx++
			}
		case "team":
			if form.teamIdx < len(form.teams)-1 {
				form.teamIdx++
			}
		}

	case "enter":
		// Check if we're on the submit button first
		if form.cursor == len(form.fields) {
			if len(form.users) == 0 {
				return m, func() tea.Msg {
					return errorMsg("No users available. Please add users first.")
				}
			}
			return m, m.submitRole()
		}

		// Now safe to access form.fields[form.cursor]
		if form.cursor >= len(form.fields) {
			return m, nil
		}

		field := form.fields[form.cursor]

		// Handle popup opening for relation fields
		switch field {
		case "user":
			if len(form.allUsers) > 0 {
				rows := []TableRow{}
				for _, user := range form.allUsers {
					id := extractID(user)
					// Parse user string: "Name <email> (ID)"
					parts := strings.Split(user, " <")
					name := parts[0]
					email := ""
					if len(parts) > 1 {
						emailParts := strings.Split(parts[1], "> (")
						email = emailParts[0]
					}
					rows = append(rows, TableRow{
						ID:      id,
						Columns: []string{name, email, id},
					})
				}
				m.popup = NewTablePopup("Select User", []string{"Name", "Email", "ID"}, rows)
				m.popupActive = true
				return m, nil
			}

		case "assigned_by":
			if len(form.allAssigners) > 0 {
				rows := []TableRow{}
				for _, assigner := range form.allAssigners {
					id := extractID(assigner)
					parts := strings.Split(assigner, " <")
					name := parts[0]
					email := ""
					if len(parts) > 1 {
						emailParts := strings.Split(parts[1], "> (")
						email = emailParts[0]
					}
					rows = append(rows, TableRow{
						ID:      id,
						Columns: []string{name, email, id},
					})
				}
				m.popup = NewTablePopup("Select Assigner", []string{"Name", "Email", "ID"}, rows)
				m.popupActive = true
				return m, nil
			}

		case "church":
			if len(form.churches) > 1 { // More than just "(none)"
				rows := []TableRow{}
				for _, church := range form.churches {
					if church == "(none)" {
						continue
					}
					id := extractID(church)
					name := church
					if idx := strings.LastIndex(church, "("); idx != -1 {
						name = strings.TrimSpace(church[:idx])
					}
					rows = append(rows, TableRow{
						ID:      id,
						Columns: []string{name, id},
					})
				}
				m.popup = NewTablePopup("Select Church", []string{"Name", "ID"}, rows)
				m.popupActive = true
				return m, nil
			}

		case "project":
			if len(form.projects) > 1 { // More than just "(none)"
				rows := []TableRow{}
				for _, project := range form.projects {
					if project == "(none)" {
						continue
					}
					id := extractID(project)
					name := project
					if idx := strings.LastIndex(project, "("); idx != -1 {
						name = strings.TrimSpace(project[:idx])
					}
					rows = append(rows, TableRow{
						ID:      id,
						Columns: []string{name, id},
					})
				}
				m.popup = NewTablePopup("Select Project", []string{"Name", "ID"}, rows)
				m.popupActive = true
				return m, nil
			}

		case "team":
			if len(form.teams) > 1 { // More than just "(none)"
				rows := []TableRow{}
				for _, team := range form.teams {
					if team == "(none)" {
						continue
					}
					id := extractID(team)
					name := team
					if idx := strings.LastIndex(team, "("); idx != -1 {
						name = strings.TrimSpace(team[:idx])
					}
					rows = append(rows, TableRow{
						ID:      id,
						Columns: []string{name, id},
					})
				}
				m.popup = NewTablePopup("Select Team", []string{"Name", "ID"}, rows)
				m.popupActive = true
				return m, nil
			}
		}

	default:
		// Handle text input for filtering user and assigner fields
		if form.cursor < len(form.fields) {
			field := form.fields[form.cursor]
			if (field == "user" || field == "assigned_by") && len(msg.String()) == 1 {
				char := msg.String()
				if field == "user" {
					form.userFilter += char
					form.applyUserFilter()
				} else if field == "assigned_by" {
					form.assignerFilter += char
					form.applyAssignerFilter()
				}
			}
		}
	}

	return m, nil
}

func (m model) submitRole() tea.Cmd {
	return func() tea.Msg {
		form := m.roleForm

		id := ulid.NewUserRoleID()
		userID := form.getSelectedUserID()
		role := form.roles[form.roleIdx]
		assignedBy := form.getSelectedAssignerID()
		churchID := form.getSelectedChurchID()
		projectID := form.getSelectedProjectID()
		teamID := form.getSelectedTeamID()

		if userID == "" {
			return errorMsg("No user selected")
		}
		if assignedBy == "" {
			return errorMsg("No assigner selected")
		}

		// Validate scope based on role
		switch role {
		case "SUPERADMIN", "ADMIN", "USER", "M2M":
			if churchID != nil || projectID != nil || teamID != nil {
				return errorMsg(fmt.Sprintf("%s role must not have any scope", role))
			}
		case "CHURCH_ADMIN":
			if churchID == nil {
				return errorMsg("CHURCH_ADMIN role requires church scope")
			}
			if projectID != nil || teamID != nil {
				return errorMsg("CHURCH_ADMIN role must have only church scope")
			}
		case "PROJECT_ADMIN":
			if projectID == nil {
				return errorMsg("PROJECT_ADMIN role requires project scope")
			}
			if churchID != nil || teamID != nil {
				return errorMsg("PROJECT_ADMIN role must have only project scope")
			}
		case "TEAM_LEAD":
			if teamID == nil {
				return errorMsg("TEAM_LEAD role requires team scope")
			}
			if churchID != nil || projectID != nil {
				return errorMsg("TEAM_LEAD role must have only team scope")
			}
		}

		query := `
			INSERT INTO user_roles (id, user_id, role, church_id, project_id, team_id, assigned_by, assigned_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`

		assignedAt := pgtype.Timestamptz{}
		err := assignedAt.Scan(time.Now())
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to convert timestamp: %v", err))
		}

		ctx := context.Background()
		_, err = m.db.Pool.Exec(ctx, query, id, userID, role, churchID, projectID, teamID, assignedBy, assignedAt)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to assign role: %v", err))
		}

		scopeStr := "global"
		if churchID != nil {
			scopeStr = fmt.Sprintf("church: %s", *churchID)
		} else if projectID != nil {
			scopeStr = fmt.Sprintf("project: %s", *projectID)
		} else if teamID != nil {
			scopeStr = fmt.Sprintf("team: %s", *teamID)
		}

		return successMsg(fmt.Sprintf("Role %s assigned successfully! ID: %s (scope: %s)", role, id, scopeStr))
	}
}

func (m model) viewRoleForm() string {
	var s strings.Builder
	form := m.roleForm

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	s.WriteString(labelStyle.Render("Assign User Role"))
	s.WriteString("\n\n")

	if form.loadingMsg != "" {
		s.WriteString(warningStyle.Render(form.loadingMsg))
		s.WriteString("\n\n")
	}

	// Role field
	cursor := " "
	if form.cursor == 0 {
		cursor = ">"
	}
	roleStr := form.roles[form.roleIdx]
	s.WriteString(fmt.Sprintf("%s Role: %s\n", cursor, selectedStyle.Render(roleStr)))
	if form.scopeHelp != "" {
		s.WriteString("  " + hintStyle.Render(form.scopeHelp) + "\n")
	}

	// User field
	cursor = " "
	if form.cursor == 1 {
		cursor = ">"
	}
	userStr := ""
	if len(form.users) == 0 {
		userStr = warningStyle.Render("[loading...]")
	} else {
		userStr = form.users[form.userIdx]
	}
	s.WriteString(fmt.Sprintf("%s User: %s\n", cursor, userStr))
	if form.userFilter != "" {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Italic(true)
		s.WriteString("  " + filterStyle.Render(fmt.Sprintf("Filter: '%s' (%d/%d matches)", form.userFilter, len(form.users), len(form.allUsers))) + "\n")
	} else if form.cursor == 1 {
		s.WriteString("  " + hintStyle.Render("Type to filter by name or email, Ctrl+D to clear") + "\n")
	}

	// Assigned by field
	cursor = " "
	if form.cursor == 2 {
		cursor = ">"
	}
	assignerStr := ""
	if len(form.assigners) == 0 {
		assignerStr = warningStyle.Render("[loading...]")
	} else {
		assignerStr = form.assigners[form.assignerIdx]
	}
	s.WriteString(fmt.Sprintf("%s Assigned By: %s\n", cursor, assignerStr))
	if form.assignerFilter != "" {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Italic(true)
		s.WriteString("  " + filterStyle.Render(fmt.Sprintf("Filter: '%s' (%d/%d matches)", form.assignerFilter, len(form.assigners), len(form.allAssigners))) + "\n")
	} else if form.cursor == 2 {
		s.WriteString("  " + hintStyle.Render("Type to filter by name or email, Ctrl+D to clear") + "\n")
	}

	s.WriteString("\n" + hintStyle.Render("--- Scope (optional, based on role) ---") + "\n")

	// Church field
	cursor = " "
	if form.cursor == 3 {
		cursor = ">"
	}
	churchStr := form.churches[form.churchIdx]
	s.WriteString(fmt.Sprintf("%s Church: %s\n", cursor, churchStr))

	// Project field
	cursor = " "
	if form.cursor == 4 {
		cursor = ">"
	}
	projectStr := form.projects[form.projectIdx]
	s.WriteString(fmt.Sprintf("%s Project: %s\n", cursor, projectStr))

	// Team field
	cursor = " "
	if form.cursor == 5 {
		cursor = ">"
	}
	teamStr := form.teams[form.teamIdx]
	s.WriteString(fmt.Sprintf("%s Team: %s\n", cursor, teamStr))

	// Submit button
	cursor = " "
	if form.cursor == len(form.fields) {
		cursor = ">"
	}
	s.WriteString(fmt.Sprintf("\n%s Submit\n", cursor))

	s.WriteString("\n")
	s.WriteString(hintStyle.Render("Use arrow keys to navigate, type to filter users, Enter to open table view or submit, Esc to cancel"))
	s.WriteString("\n")
	s.WriteString(hintStyle.Render("Global roles (SUPERADMIN, ADMIN, USER, M2M): no scope"))
	s.WriteString("\n")
	s.WriteString(hintStyle.Render("CHURCH_ADMIN: church scope only, PROJECT_ADMIN: project only, TEAM_LEAD: team only"))

	return s.String()
}
