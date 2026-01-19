package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TableRow represents a row in the table popup
type TableRow struct {
	ID      string
	Columns []string // Display columns
}

// TablePopup represents a popup with a filterable table
type TablePopup struct {
	title       string
	headers     []string
	allRows     []TableRow
	rows        []TableRow
	cursor      int
	filter      string
	columnWidth []int
}

// NewTablePopup creates a new table popup
func NewTablePopup(title string, headers []string, rows []TableRow) *TablePopup {
	// Calculate column widths
	columnWidth := make([]int, len(headers))
	for i, header := range headers {
		columnWidth[i] = len(header)
	}
	for _, row := range rows {
		for i, col := range row.Columns {
			if i < len(columnWidth) && len(col) > columnWidth[i] {
				columnWidth[i] = len(col)
			}
		}
	}

	return &TablePopup{
		title:       title,
		headers:     headers,
		allRows:     rows,
		rows:        rows,
		cursor:      0,
		columnWidth: columnWidth,
	}
}

// Update handles key events for the table popup
func (t *TablePopup) Update(msg tea.KeyMsg) (closed bool, selected *TableRow) {
	switch msg.String() {
	case "esc", "q":
		return true, nil

	case "up":
		if t.cursor > 0 {
			t.cursor--
		}

	case "down":
		if t.cursor < len(t.rows)-1 {
			t.cursor++
		}

	case "enter":
		if len(t.rows) > 0 {
			return true, &t.rows[t.cursor]
		}
		return true, nil

	case "backspace":
		if len(t.filter) > 0 {
			t.filter = t.filter[:len(t.filter)-1]
			t.applyFilter()
		}

	case "ctrl+d":
		t.filter = ""
		t.applyFilter()

	default:
		// Handle text input for filtering
		if len(msg.String()) == 1 {
			t.filter += msg.String()
			t.applyFilter()
		}
	}

	return false, nil
}

// applyFilter filters rows based on the filter string
func (t *TablePopup) applyFilter() {
	if t.filter == "" {
		t.rows = t.allRows
		if t.cursor >= len(t.rows) {
			t.cursor = len(t.rows) - 1
		}
		if t.cursor < 0 {
			t.cursor = 0
		}
		return
	}

	filtered := []TableRow{}
	filter := strings.ToLower(t.filter)
	for _, row := range t.allRows {
		// Search across all columns
		found := false
		for _, col := range row.Columns {
			if strings.Contains(strings.ToLower(col), filter) {
				found = true
				break
			}
		}
		if found {
			filtered = append(filtered, row)
		}
	}

	if len(filtered) > 0 {
		t.rows = filtered
		if t.cursor >= len(filtered) {
			t.cursor = len(filtered) - 1
		}
	} else {
		t.rows = t.allRows
	}
}

// View renders the table popup
func (t *TablePopup) View() string {
	var s strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("11")).
		Background(lipgloss.Color("237"))

	selectedRowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("12"))

	normalRowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")).
		Padding(1, 2)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)

	filterStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Bold(true)

	// Build table content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render(t.title))
	content.WriteString("\n\n")

	// Filter info
	if t.filter != "" {
		content.WriteString(filterStyle.Render(fmt.Sprintf("Filter: '%s' ", t.filter)))
		content.WriteString(hintStyle.Render(fmt.Sprintf("(%d/%d matches)", len(t.rows), len(t.allRows))))
		content.WriteString("\n\n")
	}

	// Headers
	headerRow := ""
	for i, header := range t.headers {
		width := t.columnWidth[i]
		if width > 40 {
			width = 40 // Max column width
		}
		headerRow += fmt.Sprintf("%-*s", width+2, header)
	}
	content.WriteString(headerStyle.Render(headerRow))
	content.WriteString("\n")

	// Separator
	separator := ""
	for _, width := range t.columnWidth {
		w := width
		if w > 40 {
			w = 40
		}
		separator += strings.Repeat("─", w+2)
	}
	content.WriteString(hintStyle.Render(separator))
	content.WriteString("\n")

	// Rows (show max 15 rows)
	startIdx := 0
	endIdx := len(t.rows)

	if len(t.rows) > 15 {
		// Show rows around cursor
		startIdx = t.cursor - 7
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + 15
		if endIdx > len(t.rows) {
			endIdx = len(t.rows)
			startIdx = endIdx - 15
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	for i := startIdx; i < endIdx; i++ {
		row := t.rows[i]
		rowStr := ""
		for j, col := range row.Columns {
			width := t.columnWidth[j]
			if width > 40 {
				width = 40
			}
			// Truncate if too long
			displayCol := col
			if len(col) > width {
				displayCol = col[:width-3] + "..."
			}
			rowStr += fmt.Sprintf("%-*s", width+2, displayCol)
		}

		if i == t.cursor {
			content.WriteString(selectedRowStyle.Render("> " + rowStr))
		} else {
			content.WriteString(normalRowStyle.Render("  " + rowStr))
		}
		content.WriteString("\n")
	}

	// Show scroll indicator
	if len(t.rows) > 15 {
		content.WriteString("\n")
		content.WriteString(hintStyle.Render(fmt.Sprintf("Showing %d-%d of %d rows", startIdx+1, endIdx, len(t.rows))))
		content.WriteString("\n")
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(hintStyle.Render("Type to filter • ↑↓ Navigate • Enter to select • Esc to cancel • Ctrl+D to clear filter"))

	// Wrap in border
	s.WriteString(borderStyle.Render(content.String()))

	return s.String()
}
