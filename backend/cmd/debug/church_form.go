package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/bcc-media/wayfarer/internal/ulid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaswdr/faker"
)

type churchFormModel struct {
	fields      []string
	cursor      int
	values      map[string]string
	categories  []string
	categoryIdx int
	fake        faker.Faker
}

func newChurchForm() *churchFormModel {
	fake := faker.NewWithSeed(rand.NewSource(rand.Int63()))

	return &churchFormModel{
		fields:      []string{"name", "country", "category"},
		cursor:      0,
		values:      make(map[string]string),
		categories:  []string{"S", "L", "XL"},
		categoryIdx: 1, // Default to L
		fake:        fake,
	}
}

func (f *churchFormModel) generateDefaults() {
	if f.values["name"] == "" {
		f.values["name"] = f.fake.Company().Name() + " Church"
	}
	if f.values["country"] == "" {
		f.values["country"] = f.fake.Address().Country()
	}
	f.values["category"] = f.categories[f.categoryIdx]
}

func (m model) updateChurchForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := m.churchForm

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

	case "left":
		if form.fields[form.cursor] == "category" && form.categoryIdx > 0 {
			form.categoryIdx--
		}

	case "right":
		if form.fields[form.cursor] == "category" && form.categoryIdx < len(form.categories)-1 {
			form.categoryIdx++
		}

	case "enter":
		if form.cursor == len(form.fields) { // Submit button
			form.generateDefaults()
			return m, m.submitChurch()
		}

	default:
		// Handle text input for name and country
		if form.cursor < 2 {
			field := form.fields[form.cursor]
			if msg.String() == "backspace" {
				if len(form.values[field]) > 0 {
					form.values[field] = form.values[field][:len(form.values[field])-1]
				}
			} else if len(msg.String()) == 1 {
				form.values[field] += msg.String()
			}
		}
	}

	return m, nil
}

func (m model) submitChurch() tea.Cmd {
	return func() tea.Msg {
		form := m.churchForm
		form.generateDefaults()

		id := ulid.NewChurchID()
		name := form.values["name"]
		country := form.values["country"]
		category := form.categories[form.categoryIdx]

		query := `
			INSERT INTO churches (id, name, country, category)
			VALUES ($1, $2, $3, $4)
		`

		ctx := context.Background()
		_, err := m.db.Pool.Exec(ctx, query, id, name, country, category)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to create church: %v", err))
		}

		return successMsg(fmt.Sprintf("Church created successfully! ID: %s", id))
	}
}

func (m model) viewChurchForm() string {
	var s strings.Builder
	form := m.churchForm

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	s.WriteString(labelStyle.Render("Add Church"))
	s.WriteString("\n\n")

	// Name field
	cursor := " "
	if form.cursor == 0 {
		cursor = ">"
	}
	nameValue := form.values["name"]
	if nameValue == "" {
		nameValue = hintStyle.Render("[auto-generated]")
	}
	s.WriteString(fmt.Sprintf("%s Name: %s\n", cursor, nameValue))

	// Country field
	cursor = " "
	if form.cursor == 1 {
		cursor = ">"
	}
	countryValue := form.values["country"]
	if countryValue == "" {
		countryValue = hintStyle.Render("[auto-generated]")
	}
	s.WriteString(fmt.Sprintf("%s Country: %s\n", cursor, countryValue))

	// Category field
	cursor = " "
	if form.cursor == 2 {
		cursor = ">"
	}
	categoryStr := ""
	for i, cat := range form.categories {
		if i == form.categoryIdx {
			categoryStr += selectedStyle.Render(fmt.Sprintf("[%s]", cat))
		} else {
			categoryStr += fmt.Sprintf(" %s ", cat)
		}
		if i < len(form.categories)-1 {
			categoryStr += " "
		}
	}
	s.WriteString(fmt.Sprintf("%s Category: %s\n", cursor, categoryStr))

	// Submit button
	cursor = " "
	if form.cursor == len(form.fields) {
		cursor = ">"
	}
	s.WriteString(fmt.Sprintf("\n%s Submit\n", cursor))

	s.WriteString("\n")
	s.WriteString(hintStyle.Render("Use arrow keys to navigate, type to enter text, Enter to submit, Esc to cancel"))
	s.WriteString("\n")
	s.WriteString(hintStyle.Render("Empty fields will be auto-generated with random data"))

	return s.String()
}
