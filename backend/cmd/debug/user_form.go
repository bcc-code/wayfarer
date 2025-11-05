package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/ulid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jaswdr/faker"
)

type userFormModel struct {
	db         *database.DB
	fields     []string
	cursor     int
	values     map[string]string
	genders    []string
	genderIdx  int
	churches   []string
	churchIdx  int
	fake       faker.Faker
	loadingMsg string
}

func newUserForm(db *database.DB) *userFormModel {
	fake := faker.NewWithSeed(rand.NewSource(rand.Int63()))

	form := &userFormModel{
		db:        db,
		fields:    []string{"name", "email", "members_id", "gender", "church", "avatar_url"},
		cursor:    0,
		values:    make(map[string]string),
		genders:   []string{"MALE", "FEMALE"},
		genderIdx: 0,
		churches:  []string{},
		churchIdx: 0,
		fake:      fake,
	}

	// Load churches in background
	go form.loadChurches()

	return form
}

func (f *userFormModel) loadChurches() {
	ctx := context.Background()
	query := `SELECT id, name FROM churches ORDER BY name`

	rows, err := f.db.Pool.Query(ctx, query)
	if err != nil {
		f.loadingMsg = fmt.Sprintf("Error loading churches: %v", err)
		return
	}
	defer rows.Close()

	churches := []string{}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		churches = append(churches, fmt.Sprintf("%s (%s)", name, id))
	}

	if len(churches) == 0 {
		f.loadingMsg = "No churches found. Please add a church first."
	} else {
		f.churches = churches
		f.loadingMsg = ""
	}
}

func (f *userFormModel) generateDefaults() {
	if f.values["name"] == "" {
		gender := f.genders[f.genderIdx]
		if gender == "MALE" {
			f.values["name"] = f.fake.Person().FirstNameMale() + " " + f.fake.Person().LastName()
		} else {
			f.values["name"] = f.fake.Person().FirstNameFemale() + " " + f.fake.Person().LastName()
		}
	}
	if f.values["email"] == "" {
		f.values["email"] = f.fake.Internet().Email()
	}
	if f.values["members_id"] == "" {
		f.values["members_id"] = fmt.Sprintf("MEM-%d", 10000+rand.Intn(90000))
	}
	if f.values["avatar_url"] == "" {
		f.values["avatar_url"] = fmt.Sprintf("https://i.pravatar.cc/150?img=%d", rand.Intn(70)+1)
	}
}

func (f *userFormModel) getSelectedChurchID() string {
	if len(f.churches) == 0 {
		return ""
	}
	church := f.churches[f.churchIdx]
	// Extract ID from "Name (ID)" format
	start := strings.LastIndex(church, "(")
	end := strings.LastIndex(church, ")")
	if start != -1 && end != -1 {
		return church[start+1 : end]
	}
	return ""
}

func (m model) updateUserForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := m.userForm

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
		field := form.fields[form.cursor]
		if field == "gender" && form.genderIdx > 0 {
			form.genderIdx--
		} else if field == "church" && form.churchIdx > 0 {
			form.churchIdx--
		}

	case "right":
		field := form.fields[form.cursor]
		if field == "gender" && form.genderIdx < len(form.genders)-1 {
			form.genderIdx++
		} else if field == "church" && form.churchIdx < len(form.churches)-1 {
			form.churchIdx++
		}

	case "enter":
		if form.cursor == len(form.fields) { // Submit button
			if len(form.churches) == 0 {
				return m, func() tea.Msg {
					return errorMsg("No churches available. Please add a church first.")
				}
			}
			form.generateDefaults()
			return m, m.submitUser()
		} else if form.cursor < len(form.fields) {
			field := form.fields[form.cursor]
			if field == "church" {
				// Open popup for church selection
				if len(form.churches) > 1 { // More than just "(none)"
					rows := []TableRow{}
					for _, church := range form.churches {
						if church == "(none)" {
							continue
						}
						id := extractID(church)
						// Parse church string: "Name (ID)"
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
			}
		}

	default:
		// Handle text input for text fields
		if form.cursor < len(form.fields) {
			field := form.fields[form.cursor]
			if field != "gender" && field != "church" {
				if msg.String() == "backspace" {
					if len(form.values[field]) > 0 {
						form.values[field] = form.values[field][:len(form.values[field])-1]
					}
				} else if len(msg.String()) == 1 {
					form.values[field] += msg.String()
				}
			}
		}
	}

	return m, nil
}

func (m model) submitUser() tea.Cmd {
	return func() tea.Msg {
		form := m.userForm
		form.generateDefaults()

		id := ulid.NewUserID()
		name := form.values["name"]
		email := form.values["email"]
		membersID := form.values["members_id"]
		gender := form.genders[form.genderIdx]
		avatarURL := form.values["avatar_url"]
		churchID := form.getSelectedChurchID()

		if churchID == "" {
			return errorMsg("No church selected")
		}

		// Generate random birthdate (age 13-80)
		age := 13 + rand.Intn(68)
		now := time.Now()
		birthdate := now.AddDate(-age, -rand.Intn(12), -rand.Intn(28))

		birthdatePg := pgtype.Date{}
		err := birthdatePg.Scan(birthdate)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to convert birthdate: %v", err))
		}

		query := `
			INSERT INTO users (id, members_id, email, name, gender, birthdate, church_id, avatar_url)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`

		ctx := context.Background()
		_, err = m.db.Pool.Exec(ctx, query, id, membersID, email, name, gender, birthdatePg, churchID, avatarURL)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to create user: %v", err))
		}

		return successMsg(fmt.Sprintf("User created successfully! ID: %s", id))
	}
}

func (m model) viewUserForm() string {
	var s strings.Builder
	form := m.userForm

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	s.WriteString(labelStyle.Render("Add User"))
	s.WriteString("\n\n")

	if form.loadingMsg != "" {
		s.WriteString(warningStyle.Render(form.loadingMsg))
		s.WriteString("\n\n")
	}

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

	// Email field
	cursor = " "
	if form.cursor == 1 {
		cursor = ">"
	}
	emailValue := form.values["email"]
	if emailValue == "" {
		emailValue = hintStyle.Render("[auto-generated]")
	}
	s.WriteString(fmt.Sprintf("%s Email: %s\n", cursor, emailValue))

	// Members ID field
	cursor = " "
	if form.cursor == 2 {
		cursor = ">"
	}
	membersIDValue := form.values["members_id"]
	if membersIDValue == "" {
		membersIDValue = hintStyle.Render("[auto-generated]")
	}
	s.WriteString(fmt.Sprintf("%s Members ID: %s\n", cursor, membersIDValue))

	// Gender field
	cursor = " "
	if form.cursor == 3 {
		cursor = ">"
	}
	genderStr := ""
	for i, gen := range form.genders {
		if i == form.genderIdx {
			genderStr += selectedStyle.Render(fmt.Sprintf("[%s]", gen))
		} else {
			genderStr += fmt.Sprintf(" %s ", gen)
		}
		if i < len(form.genders)-1 {
			genderStr += " "
		}
	}
	s.WriteString(fmt.Sprintf("%s Gender: %s\n", cursor, genderStr))

	// Church field
	cursor = " "
	if form.cursor == 4 {
		cursor = ">"
	}
	churchStr := ""
	if len(form.churches) == 0 {
		churchStr = warningStyle.Render("[loading...]")
	} else {
		churchStr = form.churches[form.churchIdx]
	}
	s.WriteString(fmt.Sprintf("%s Church: %s\n", cursor, churchStr))

	// Avatar URL field
	cursor = " "
	if form.cursor == 5 {
		cursor = ">"
	}
	avatarValue := form.values["avatar_url"]
	if avatarValue == "" {
		avatarValue = hintStyle.Render("[auto-generated]")
	}
	s.WriteString(fmt.Sprintf("%s Avatar URL: %s\n", cursor, avatarValue))

	// Submit button
	cursor = " "
	if form.cursor == len(form.fields) {
		cursor = ">"
	}
	s.WriteString(fmt.Sprintf("\n%s Submit\n", cursor))

	s.WriteString("\n")
	s.WriteString(hintStyle.Render("Use arrow keys to navigate, type to enter text, Enter on church to open table view, Esc to cancel"))
	s.WriteString("\n")
	s.WriteString(hintStyle.Render("Empty fields will be auto-generated with random data"))
	s.WriteString("\n")
	s.WriteString(hintStyle.Render("Birthdate is auto-generated (age 13-80)"))

	return s.String()
}
