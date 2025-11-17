package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/bcc-media/wayfarer/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/golang-jwt/jwt/v5"
)

type WayfarerClaims struct {
	UserID    string   `json:"user_id"`
	UserRoles []string `json:"user_roles"`
	jwt.RegisteredClaims
}

type tokenFormModel struct {
	db               *database.DB
	searchQuery      string
	selectedUserID   string
	selectedUserName string
	generatedToken   string
}

func newTokenForm(db *database.DB) *tokenFormModel {
	return &tokenFormModel{
		db:          db,
		searchQuery: "",
	}
}

func (m model) updateTokenForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := m.tokenForm

	switch msg.String() {
	case "ctrl+c", "esc":
		m.screen = ScreenMenu
		m.cursor = 0
		return m, nil

	case "backspace":
		if len(form.searchQuery) > 0 {
			form.searchQuery = form.searchQuery[:len(form.searchQuery)-1]
		}

	case "enter":
		if form.searchQuery != "" {
			return m, m.searchUsers()
		}

	default:
		if len(msg.String()) == 1 {
			form.searchQuery += msg.String()
		}
	}

	return m, nil
}

func (m model) searchUsers() tea.Cmd {
	return func() tea.Msg {
		form := m.tokenForm
		ctx := context.Background()

		query := `
			SELECT id, name, email, members_id, church_id
			FROM users
			WHERE name ILIKE $1 OR email ILIKE $1 OR members_id ILIKE $1
			ORDER BY name
			LIMIT 50
		`

		searchPattern := "%" + form.searchQuery + "%"
		rows, err := m.db.Pool.Query(ctx, query, searchPattern)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to search users: %v", err))
		}
		defer rows.Close()

		results := []TableRow{}
		for rows.Next() {
			var id, name, email, membersID, churchID string
			if err := rows.Scan(&id, &name, &email, &membersID, &churchID); err != nil {
				continue
			}
			results = append(results, TableRow{
				ID:      id,
				Columns: []string{name, email, membersID, id},
			})
		}

		if len(results) == 0 {
			return errorMsg("No users found matching your search")
		}

		// Update model with popup
		return showPopupMsg{
			title:   "Select User",
			headers: []string{"Name", "Email", "Members ID", "User ID"},
			rows:    results,
		}
	}
}

func (m model) generateToken() tea.Cmd {
	return func() tea.Msg {
		form := m.tokenForm
		secret := "your-secret-key-for-signing-wayfarer-jwts"

		now := time.Now()
		claims := WayfarerClaims{
			UserID:    form.selectedUserID,
			UserRoles: []string{"user"},
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "wayfarer",
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to generate token: %v", err))
		}

		form.generatedToken = tokenString

		// Copy to clipboard
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(tokenString)
		if err := cmd.Run(); err != nil {
			return successMsg(fmt.Sprintf("Token generated for user: %s (clipboard copy failed)", form.selectedUserName))
		}

		return successMsg(fmt.Sprintf("Token generated and copied to clipboard for user: %s", form.selectedUserName))
	}
}

func (m model) viewTokenForm() string {
	var s strings.Builder
	form := m.tokenForm

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	tokenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Background(lipgloss.Color("0")).Padding(1)

	s.WriteString(labelStyle.Render("Search User & Generate Token"))
	s.WriteString("\n\n")

	s.WriteString("Search: ")
	s.WriteString(form.searchQuery)
	s.WriteString("_")
	s.WriteString("\n\n")

	s.WriteString(hintStyle.Render("Type to search by name, email, or members ID, then press Enter"))
	s.WriteString("\n")

	if form.selectedUserID != "" {
		s.WriteString("\n")
		s.WriteString(labelStyle.Render("Selected User:"))
		s.WriteString(fmt.Sprintf(" %s (ID: %s)", form.selectedUserName, form.selectedUserID))
		s.WriteString("\n\n")
	}

	if form.generatedToken != "" {
		s.WriteString(labelStyle.Render("Generated Token:"))
		s.WriteString("\n\n")
		s.WriteString(tokenStyle.Render(form.generatedToken))
		s.WriteString("\n\n")
		s.WriteString(hintStyle.Render("Token has been copied to clipboard!"))
		s.WriteString("\n")
		s.WriteString(hintStyle.Render("Token is valid for 24 hours"))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(hintStyle.Render("Press Esc to return to menu"))

	return s.String()
}

type showPopupMsg struct {
	title   string
	headers []string
	rows    []TableRow
}
