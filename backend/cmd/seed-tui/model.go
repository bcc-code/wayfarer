package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// readLastLines reads the last n lines from a file
func readLastLines(filename string, n int) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	return lines, scanner.Err()
}

// Screen represents different screens in the TUI
type Screen int

const (
	ScreenSelectProject Screen = iota
	ScreenSelectAction
	ScreenConfig
	ScreenConfirm
	ScreenRunning
	ScreenComplete
)

// Action represents the action to perform
type Action int

const (
	ActionSeedTeams Action = iota
	ActionReassignTeamLeads
)

// SeedConfig holds the configuration for seeding
type SeedConfig struct {
	TeamCount int
	TeamSize  int
	MinPoints int
	MaxPoints int
}

// model is the main TUI model
type model struct {
	ctx              context.Context
	db               *database.DB
	screen           Screen
	projects         []*sqlc.GetProjectsForSeedRow
	projectCursor    int
	selectedProject  *sqlc.GetProjectsForSeedRow
	selectedAction   Action
	actionCursor     int
	config           SeedConfig
	configCursor     int
	configEditing    bool
	configEditBuffer string
	availableUsers   int
	usersNeeded      int
	usersToCreate    int
	seeder           *TeamSeeder
	progress         SeedProgress
	result           SeedResult
	err              error
	message          string
	messageStyle     lipgloss.Style
	debugLogs        []string
}

// SeedProgress tracks seeding progress
type SeedProgress struct {
	Stage           string
	Current         int
	Total           int
	UsersCreated    int
	TeamsCreated    int
	MembersAssigned int
	PointsGenerated int
}

// SeedResult holds the final result
type SeedResult struct {
	UsersCreated      int
	TeamsCreated      int
	MembersAssigned   int
	TeamLeadsAssigned int
	PointsGenerated   int
}

// Messages
type projectsLoadedMsg struct {
	projects []*sqlc.GetProjectsForSeedRow
}

type availableUsersMsg struct {
	count int
}

type seedProgressMsg struct {
	progress SeedProgress
}

type seedCompleteMsg struct {
	result SeedResult
}

type reassignTeamLeadsCompleteMsg struct {
	leadsAssigned int
}

type seedErrorMsg struct {
	err error
}

type debugLogMsg struct {
	msg string
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// initialModel creates the initial model
func initialModel(ctx context.Context, db *database.DB) model {
	return model{
		ctx:    ctx,
		db:     db,
		screen: ScreenSelectProject,
		config: SeedConfig{
			TeamCount: 300,
			TeamSize:  8,
			MinPoints: 100,
			MaxPoints: 15000,
		},
		messageStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
	}
}

// Init initializes the model and loads projects
func (m model) Init() tea.Cmd {
	return m.loadProjects()
}

// loadProjects fetches projects from the database
func (m model) loadProjects() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.db.Queries.GetProjectsForSeed(m.ctx)
		if err != nil {
			return seedErrorMsg{err: err}
		}
		return projectsLoadedMsg{projects: projects}
	}
}

// loadAvailableUsers counts users not in teams for the selected project
func (m model) loadAvailableUsers() tea.Cmd {
	return func() tea.Msg {
		if m.selectedProject == nil {
			return seedErrorMsg{err: fmt.Errorf("no project selected")}
		}
		count, err := m.db.Queries.GetUserCountNotInTeamForProject(m.ctx, m.selectedProject.ID)
		if err != nil {
			return seedErrorMsg{err: err}
		}
		return availableUsersMsg{count: int(count)}
	}
}

// startSeeding begins the seeding process
func (m model) startSeeding() tea.Cmd {
	return func() tea.Msg {
		if m.selectedProject == nil {
			return seedErrorMsg{err: fmt.Errorf("no project selected")}
		}

		seeder := NewTeamSeeder(m.ctx, m.db, m.selectedProject.ID, m.config)

		// Run seeding synchronously (bubbletea runs Cmd in a goroutine)
		result, err := seeder.RunSync()
		if err != nil {
			return seedErrorMsg{err: err}
		}

		return seedCompleteMsg{result: result}
	}
}

// startReassignTeamLeads begins the team lead reassignment process
func (m model) startReassignTeamLeads() tea.Cmd {
	return func() tea.Msg {
		if m.selectedProject == nil {
			return seedErrorMsg{err: fmt.Errorf("no project selected")}
		}

		seeder := NewTeamSeeder(m.ctx, m.db, m.selectedProject.ID, m.config)

		leadsAssigned, err := seeder.ReassignTeamLeadsRandomly()
		if err != nil {
			return seedErrorMsg{err: err}
		}

		return reassignTeamLeadsCompleteMsg{leadsAssigned: leadsAssigned}
	}
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case projectsLoadedMsg:
		m.projects = msg.projects
		return m, nil

	case availableUsersMsg:
		m.availableUsers = msg.count
		m.updateUsersNeeded()
		return m, nil

	case seedProgressMsg:
		m.progress = msg.progress
		return m, nil

	case seedCompleteMsg:
		m.result = msg.result
		m.screen = ScreenComplete
		return m, nil

	case reassignTeamLeadsCompleteMsg:
		m.result = SeedResult{TeamLeadsAssigned: msg.leadsAssigned}
		m.screen = ScreenComplete
		return m, nil

	case seedErrorMsg:
		m.err = msg.err
		m.message = fmt.Sprintf("Error: %v", msg.err)
		m.messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		m.screen = ScreenComplete
		return m, nil

	case tickMsg:
		// Refresh the view (for log display) while on running screen
		if m.screen == ScreenRunning {
			return m, tickCmd()
		}
		return m, nil
	}

	return m, nil
}

// handleKeyMsg handles keyboard input
func (m model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.screen {
	case ScreenSelectProject:
		return m.updateProjectSelect(msg)
	case ScreenSelectAction:
		return m.updateActionSelect(msg)
	case ScreenConfig:
		return m.updateConfig(msg)
	case ScreenConfirm:
		return m.updateConfirm(msg)
	case ScreenRunning:
		// No input during running
		return m, nil
	case ScreenComplete:
		return m.updateComplete(msg)
	}

	return m, nil
}

// updateProjectSelect handles input on the project selection screen
func (m model) updateProjectSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		if m.projectCursor > 0 {
			m.projectCursor--
		}
	case "down", "j":
		if m.projectCursor < len(m.projects)-1 {
			m.projectCursor++
		}
	case "enter":
		if len(m.projects) > 0 {
			m.selectedProject = m.projects[m.projectCursor]
			m.screen = ScreenSelectAction
			m.actionCursor = 0
		}
	}
	return m, nil
}

// updateActionSelect handles input on the action selection screen
func (m model) updateActionSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.screen = ScreenSelectProject
	case "up", "k":
		if m.actionCursor > 0 {
			m.actionCursor--
		}
	case "down", "j":
		if m.actionCursor < 1 { // 2 actions: 0 and 1
			m.actionCursor++
		}
	case "enter":
		m.selectedAction = Action(m.actionCursor)
		if m.selectedAction == ActionSeedTeams {
			m.screen = ScreenConfig
			return m, m.loadAvailableUsers()
		} else if m.selectedAction == ActionReassignTeamLeads {
			m.screen = ScreenRunning
			return m, tea.Batch(m.startReassignTeamLeads(), tickCmd())
		}
	}
	return m, nil
}

// updateConfig handles input on the config screen
func (m model) updateConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.configEditing {
		return m.updateConfigEditing(msg)
	}

	switch msg.String() {
	case "q", "esc":
		m.screen = ScreenSelectProject
	case "up", "k":
		if m.configCursor > 0 {
			m.configCursor--
		}
	case "down", "j":
		if m.configCursor < 3 {
			m.configCursor++
		}
	case "enter":
		m.configEditing = true
		switch m.configCursor {
		case 0:
			m.configEditBuffer = strconv.Itoa(m.config.TeamCount)
		case 1:
			m.configEditBuffer = strconv.Itoa(m.config.TeamSize)
		case 2:
			m.configEditBuffer = strconv.Itoa(m.config.MinPoints)
		case 3:
			m.configEditBuffer = strconv.Itoa(m.config.MaxPoints)
		}
	case "tab":
		m.screen = ScreenConfirm
	}
	return m, nil
}

// updateConfigEditing handles input when editing a config value
func (m model) updateConfigEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.configEditing = false
		m.configEditBuffer = ""
	case "enter":
		value, err := strconv.Atoi(m.configEditBuffer)
		if err == nil && value > 0 {
			switch m.configCursor {
			case 0:
				m.config.TeamCount = value
			case 1:
				m.config.TeamSize = value
			case 2:
				m.config.MinPoints = value
			case 3:
				m.config.MaxPoints = value
			}
			m.updateUsersNeeded()
		}
		m.configEditing = false
		m.configEditBuffer = ""
	case "backspace":
		if len(m.configEditBuffer) > 0 {
			m.configEditBuffer = m.configEditBuffer[:len(m.configEditBuffer)-1]
		}
	default:
		// Only allow digits
		if len(msg.String()) == 1 && msg.String()[0] >= '0' && msg.String()[0] <= '9' {
			m.configEditBuffer += msg.String()
		}
	}
	return m, nil
}

// updateConfirm handles input on the confirm screen
func (m model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.screen = ScreenConfig
	case "enter", "y":
		m.screen = ScreenRunning
		// Start both the seeding process and the ticker for UI updates
		return m, tea.Batch(m.startSeeding(), tickCmd())
	case "n":
		m.screen = ScreenConfig
	}
	return m, nil
}

// updateComplete handles input on the complete screen
func (m model) updateComplete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "enter":
		return m, tea.Quit
	}
	return m, nil
}

// updateUsersNeeded calculates how many users are needed
func (m *model) updateUsersNeeded() {
	m.usersNeeded = m.config.TeamCount * m.config.TeamSize
	if m.availableUsers >= m.usersNeeded {
		m.usersToCreate = 0
	} else {
		m.usersToCreate = m.usersNeeded - m.availableUsers
	}
}

// View renders the UI
func (m model) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginBottom(1)

	s.WriteString(titleStyle.Render("Wayfarer Team Seeder"))
	s.WriteString("\n")

	// Show database info
	dbInfoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	s.WriteString(dbInfoStyle.Render("Database: connected"))
	s.WriteString("\n\n")

	if m.message != "" {
		s.WriteString(m.messageStyle.Render(m.message))
		s.WriteString("\n\n")
	}

	switch m.screen {
	case ScreenSelectProject:
		s.WriteString(m.viewProjectSelect())
	case ScreenSelectAction:
		s.WriteString(m.viewActionSelect())
	case ScreenConfig:
		s.WriteString(m.viewConfig())
	case ScreenConfirm:
		s.WriteString(m.viewConfirm())
	case ScreenRunning:
		s.WriteString(m.viewRunning())
	case ScreenComplete:
		s.WriteString(m.viewComplete())
	}

	return s.String()
}

// viewProjectSelect renders the project selection screen
func (m model) viewProjectSelect() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	s.WriteString(headerStyle.Render("Select a project:"))
	s.WriteString("\n\n")

	if len(m.projects) == 0 {
		s.WriteString("Loading projects...")
		return s.String()
	}

	for i, p := range m.projects {
		cursor := "  "
		if i == m.projectCursor {
			cursor = "> "
		}

		name := p.Name
		if p.StartDate.Valid {
			name += fmt.Sprintf(" (%s)", p.StartDate.Time.Format("2006-01-02"))
		}

		s.WriteString(fmt.Sprintf("%s%s\n", cursor, name))
	}

	s.WriteString("\n")
	s.WriteString("[Enter] Select  [q] Quit")

	return s.String()
}

// viewActionSelect renders the action selection screen
func (m model) viewActionSelect() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	s.WriteString(headerStyle.Render("Select an action:"))
	s.WriteString("\n")

	if m.selectedProject != nil {
		projectStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
		s.WriteString(fmt.Sprintf("Project: %s\n", projectStyle.Render(m.selectedProject.Name)))
	}
	s.WriteString("\n")

	actions := []struct {
		name        string
		description string
	}{
		{"Seed teams", "Create teams, assign users, and generate points"},
		{"Re-assign team leads", "Randomly re-assign team leads for all teams"},
	}

	for i, action := range actions {
		cursor := "  "
		if i == m.actionCursor {
			cursor = "> "
		}

		nameStyle := lipgloss.NewStyle().Bold(i == m.actionCursor)
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		s.WriteString(fmt.Sprintf("%s%s\n", cursor, nameStyle.Render(action.name)))
		s.WriteString(fmt.Sprintf("    %s\n", descStyle.Render(action.description)))
	}

	s.WriteString("\n")
	s.WriteString("[Enter] Select  [Esc] Back  [q] Quit")

	return s.String()
}

// viewConfig renders the configuration screen
func (m model) viewConfig() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	s.WriteString(headerStyle.Render("Configuration"))
	s.WriteString("\n")

	if m.selectedProject != nil {
		projectStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
		s.WriteString(fmt.Sprintf("Project: %s\n", projectStyle.Render(m.selectedProject.Name)))
	}

	availStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	s.WriteString(fmt.Sprintf("Available users: %s (not in teams)\n", availStyle.Render(fmt.Sprintf("%d", m.availableUsers))))
	s.WriteString("\n")

	labels := []string{"Teams to create:", "Team size:", "Min points/user:", "Max points/user:"}
	values := []int{m.config.TeamCount, m.config.TeamSize, m.config.MinPoints, m.config.MaxPoints}

	for i, label := range labels {
		cursor := "  "
		if i == m.configCursor {
			cursor = "> "
		}

		valueStr := fmt.Sprintf("%d", values[i])
		if m.configEditing && i == m.configCursor {
			valueStr = fmt.Sprintf("[ %s_ ]", m.configEditBuffer)
		} else {
			valueStr = fmt.Sprintf("[ %d ]", values[i])
		}

		s.WriteString(fmt.Sprintf("%s%-18s %s\n", cursor, label, valueStr))
	}

	s.WriteString("\n")

	// Show calculated values
	neededStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	s.WriteString(fmt.Sprintf("Users needed: %s\n", neededStyle.Render(fmt.Sprintf("%d", m.usersNeeded))))

	if m.usersToCreate > 0 {
		createStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		s.WriteString(fmt.Sprintf("Will use: %d existing\n", m.availableUsers))
		s.WriteString(fmt.Sprintf("Will create: %s new users\n", createStyle.Render(fmt.Sprintf("%d", m.usersToCreate))))
	} else {
		s.WriteString(fmt.Sprintf("Will use: %d existing (no new users needed)\n", m.usersNeeded))
	}

	s.WriteString("\n")

	if m.configEditing {
		s.WriteString("[Enter] Save  [Esc] Cancel")
	} else {
		s.WriteString("[Enter] Edit  [Tab] Continue  [Esc] Back")
	}

	return s.String()
}

// viewConfirm renders the confirmation screen
func (m model) viewConfirm() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	s.WriteString(headerStyle.Render("Confirm Seeding"))
	s.WriteString("\n\n")

	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	s.WriteString(warnStyle.Render("This will create:"))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("  - %d teams\n", m.config.TeamCount))
	s.WriteString(fmt.Sprintf("  - %d team memberships\n", m.usersNeeded))
	if m.usersToCreate > 0 {
		s.WriteString(fmt.Sprintf("  - %d new users\n", m.usersToCreate))
	}
	s.WriteString(fmt.Sprintf("  - %d-%d points per user\n", m.config.MinPoints, m.config.MaxPoints))

	s.WriteString("\n")
	s.WriteString(fmt.Sprintf("Project: %s\n", m.selectedProject.Name))

	s.WriteString("\n")
	confirmStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	s.WriteString(confirmStyle.Render("Proceed? [y/Enter] Yes  [n/Esc] No"))

	return s.String()
}

// viewRunning renders the progress screen
func (m model) viewRunning() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	s.WriteString(headerStyle.Render("Seeding..."))
	s.WriteString("\n\n")

	spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	if m.progress.Stage != "" {
		s.WriteString(fmt.Sprintf("%s %s\n", spinnerStyle.Render("~"), m.progress.Stage))
		if m.progress.Total > 0 {
			pct := float64(m.progress.Current) / float64(m.progress.Total) * 100
			barWidth := 30
			filled := int(float64(barWidth) * float64(m.progress.Current) / float64(m.progress.Total))
			bar := strings.Repeat("=", filled) + strings.Repeat("-", barWidth-filled)
			s.WriteString(fmt.Sprintf("  [%s] %.0f%%\n", bar, pct))
		}
	} else {
		s.WriteString(spinnerStyle.Render("Initializing..."))
	}

	s.WriteString("\n")
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	s.WriteString(statsStyle.Render(fmt.Sprintf("Users created: %d", m.progress.UsersCreated)))
	s.WriteString("\n")
	s.WriteString(statsStyle.Render(fmt.Sprintf("Teams created: %d", m.progress.TeamsCreated)))
	s.WriteString("\n")
	s.WriteString(statsStyle.Render(fmt.Sprintf("Members assigned: %d", m.progress.MembersAssigned)))
	s.WriteString("\n")
	s.WriteString(statsStyle.Render(fmt.Sprintf("Points generated: %d", m.progress.PointsGenerated)))

	// Show debug logs from file
	s.WriteString("\n\nDebug log (tail /tmp/seed-tui.log):\n")
	logStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	if logs, err := readLastLines("/tmp/seed-tui.log", 10); err == nil {
		for _, line := range logs {
			s.WriteString(logStyle.Render(line))
			s.WriteString("\n")
		}
	} else {
		s.WriteString(logStyle.Render(fmt.Sprintf("(no logs: %v)", err)))
	}

	return s.String()
}

// viewComplete renders the completion screen
func (m model) viewComplete() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))

	// Check if this was a team lead reassignment (only TeamLeadsAssigned is set)
	if m.selectedAction == ActionReassignTeamLeads {
		s.WriteString(headerStyle.Render("Team Lead Reassignment Complete!"))
		s.WriteString("\n\n")
		s.WriteString(fmt.Sprintf("Team leads assigned: %d\n", m.result.TeamLeadsAssigned))
	} else {
		s.WriteString(headerStyle.Render("Seeding Complete!"))
		s.WriteString("\n\n")
		s.WriteString(fmt.Sprintf("Users created:       %d\n", m.result.UsersCreated))
		s.WriteString(fmt.Sprintf("Teams created:       %d\n", m.result.TeamsCreated))
		s.WriteString(fmt.Sprintf("Members assigned:    %d\n", m.result.MembersAssigned))
		s.WriteString(fmt.Sprintf("Team leads assigned: %d\n", m.result.TeamLeadsAssigned))
		s.WriteString(fmt.Sprintf("Points generated:    %d\n", m.result.PointsGenerated))
	}

	s.WriteString("\n")
	s.WriteString("[Enter/q] Exit")

	return s.String()
}
