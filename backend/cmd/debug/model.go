package main

import (
	"fmt"
	"strings"

	"github.com/bcc-media/wayfarer/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen represents different screens in the TUI
type Screen int

const (
	ScreenMenu Screen = iota
	ScreenAddChurch
	ScreenAddUser
	ScreenAssignRole
	ScreenSearchUserToken
)

// Model holds the application state
type model struct {
	db           *database.DB
	screen       Screen
	cursor       int
	menuItems    []string
	churchForm   *churchFormModel
	userForm     *userFormModel
	roleForm     *roleFormModel
	tokenForm    *tokenFormModel
	message      string
	messageStyle lipgloss.Style
	popup        *TablePopup
	popupActive  bool
}

// initialModel creates the initial model
func initialModel(db *database.DB) model {
	return model{
		db:     db,
		screen: ScreenMenu,
		cursor: 0,
		menuItems: []string{
			"Add Church",
			"Add User",
			"Assign User Role",
			"Search User & Generate Token",
			"Quit",
		},
		messageStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
	}
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle popup first if active
		if m.popupActive && m.popup != nil {
			closed, selected := m.popup.Update(msg)
			if closed {
				m.popupActive = false
				if selected != nil {
					// Handle selection based on current screen
					return m.handlePopupSelection(selected)
				}
				return m, nil
			}
			return m, nil
		}

		// Normal screen handling
		switch m.screen {
		case ScreenMenu:
			return m.updateMenu(msg)
		case ScreenAddChurch:
			return m.updateChurchForm(msg)
		case ScreenAddUser:
			return m.updateUserForm(msg)
		case ScreenAssignRole:
			return m.updateRoleForm(msg)
		case ScreenSearchUserToken:
			return m.updateTokenForm(msg)
		}

	case successMsg:
		m.message = string(msg)
		m.messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		return m, nil

	case errorMsg:
		m.message = string(msg)
		m.messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		return m, nil

	case showPopupMsg:
		m.popup = NewTablePopup(msg.title, msg.headers, msg.rows)
		m.popupActive = true
		return m, nil
	}

	return m, nil
}

// handlePopupSelection handles the selection from a popup
func (m model) handlePopupSelection(selected *TableRow) (model, tea.Cmd) {
	switch m.screen {
	case ScreenAddUser:
		if m.userForm != nil {
			// Find the church by ID
			for i, church := range m.userForm.churches {
				if extractID(church) == selected.ID {
					m.userForm.churchIdx = i
					break
				}
			}
		}
	case ScreenSearchUserToken:
		if m.tokenForm != nil {
			m.tokenForm.selectedUserID = selected.ID
			m.tokenForm.selectedUserName = selected.Columns[0]
			return m, m.generateToken()
		}
	case ScreenAssignRole:
		if m.roleForm != nil {
			field := m.roleForm.fields[m.roleForm.cursor]
			switch field {
			case "user":
				for i, user := range m.roleForm.allUsers {
					if extractID(user) == selected.ID {
						m.roleForm.userIdx = i
						m.roleForm.users = m.roleForm.allUsers
						m.roleForm.userFilter = ""
						break
					}
				}
			case "assigned_by":
				for i, assigner := range m.roleForm.allAssigners {
					if extractID(assigner) == selected.ID {
						m.roleForm.assignerIdx = i
						m.roleForm.assigners = m.roleForm.allAssigners
						m.roleForm.assignerFilter = ""
						break
					}
				}
			case "church":
				for i, church := range m.roleForm.churches {
					if extractID(church) == selected.ID {
						m.roleForm.churchIdx = i
						break
					}
				}
			case "project":
				for i, project := range m.roleForm.projects {
					if extractID(project) == selected.ID {
						m.roleForm.projectIdx = i
						break
					}
				}
			case "team":
				for i, team := range m.roleForm.teams {
					if extractID(team) == selected.ID {
						m.roleForm.teamIdx = i
						break
					}
				}
			}
		}
	}
	return m, nil
}

// updateMenu handles menu updates
func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down":
		if m.cursor < len(m.menuItems)-1 {
			m.cursor++
		}

	case "enter":
		switch m.cursor {
		case 0: // Add Church
			m.screen = ScreenAddChurch
			m.churchForm = newChurchForm()
			m.message = ""
		case 1: // Add User
			m.screen = ScreenAddUser
			m.userForm = newUserForm(m.db)
			m.message = ""
		case 2: // Assign User Role
			m.screen = ScreenAssignRole
			m.roleForm = newRoleForm(m.db)
			m.message = ""
		case 3: // Search User & Generate Token
			m.screen = ScreenSearchUserToken
			m.tokenForm = newTokenForm(m.db)
			m.message = ""
		case 4: // Quit
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the UI
func (m model) View() string {
	// If popup is active, overlay it on top
	if m.popupActive && m.popup != nil {
		return m.popup.View()
	}

	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginBottom(1)

	s.WriteString(titleStyle.Render("Wayfarer Debug Tool"))
	s.WriteString("\n\n")

	if m.message != "" {
		s.WriteString(m.messageStyle.Render(m.message))
		s.WriteString("\n\n")
	}

	switch m.screen {
	case ScreenMenu:
		s.WriteString(m.viewMenu())
	case ScreenAddChurch:
		s.WriteString(m.viewChurchForm())
	case ScreenAddUser:
		s.WriteString(m.viewUserForm())
	case ScreenAssignRole:
		s.WriteString(m.viewRoleForm())
	case ScreenSearchUserToken:
		s.WriteString(m.viewTokenForm())
	}

	return s.String()
}

// viewMenu renders the main menu
func (m model) viewMenu() string {
	var s strings.Builder

	for i, item := range m.menuItems {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursor, item))
	}

	s.WriteString("\n\nUse arrow keys to navigate, Enter to select, q to quit\n")

	return s.String()
}

// Messages
type successMsg string
type errorMsg string
