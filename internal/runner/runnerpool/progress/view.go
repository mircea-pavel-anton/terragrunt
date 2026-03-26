package progress

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	statusStylePending   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // gray
	statusStyleRunning   = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
	statusStyleSucceeded = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	statusStyleFailed    = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
	statusStyleEarlyExit = lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // magenta

	unitNameStyle   = lipgloss.NewStyle().Bold(true)
	outputLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).PaddingLeft(4) //nolint:mnd
	durationStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	summaryStyle    = lipgloss.NewStyle().Bold(true).PaddingTop(1)
	changeSumStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).PaddingLeft(4) //nolint:mnd
)

// View renders the current state of all units.
func (m Model) View() string { //nolint:gocritic
	if len(m.order) == 0 {
		return ""
	}

	var b strings.Builder

	for _, path := range m.order {
		u, ok := m.units[path]
		if !ok {
			continue
		}

		b.WriteString(m.renderUnit(u))
		b.WriteByte('\n')
	}

	b.WriteString(m.renderSummary())

	return b.String()
}

// renderUnit renders a single unit's status line and optionally its output.
func (m Model) renderUnit(u *unitState) string { //nolint:gocritic
	icon := statusIcon(u.status)
	stStyle := statusStyle(u.status)
	name := unitNameStyle.Render(u.displayPath)
	dur := formatDuration(u.duration())
	durStr := durationStyle.Render(dur)

	header := fmt.Sprintf("%s %s %s", stStyle.Render(icon), name, durStr)

	if !u.isExpanded() {
		return header
	}

	var b strings.Builder

	b.WriteString(header)

	// For succeeded units with changes, show the change summary prominently.
	if u.status == StatusSucceeded && u.hasChanges && u.changeSummary != "" {
		b.WriteByte('\n')
		b.WriteString(changeSumStyle.Render(u.changeSummary))
	}

	// Show recent output lines.
	linesToShow := u.lines
	if u.status == StatusSucceeded && u.hasChanges {
		// For succeeded-with-changes, show all buffered lines for context.
		linesToShow = u.lines
	}

	for _, line := range linesToShow {
		b.WriteByte('\n')

		rendered := outputLineStyle.Render(line)

		// Truncate to terminal width if needed.
		if m.width > 0 && lipgloss.Width(rendered) > m.width {
			rendered = rendered[:m.width]
		}

		b.WriteString(rendered)
	}

	return b.String()
}

// renderSummary renders the bottom progress summary line.
func (m Model) renderSummary() string { //nolint:gocritic
	var total, pending, running, succeeded, failed, earlyExit int

	total = len(m.order)

	for _, path := range m.order {
		u, ok := m.units[path]
		if !ok {
			continue
		}

		switch u.status {
		case StatusPending:
			pending++
		case StatusRunning:
			running++
		case StatusSucceeded:
			succeeded++
		case StatusFailed:
			failed++
		case StatusEarlyExit:
			earlyExit++
		}
	}

	parts := []string{fmt.Sprintf("[%d/%d]", succeeded+failed+earlyExit, total)}

	if running > 0 {
		parts = append(parts, statusStyleRunning.Render(fmt.Sprintf("%d running", running)))
	}

	if succeeded > 0 {
		parts = append(parts, statusStyleSucceeded.Render(fmt.Sprintf("%d succeeded", succeeded)))
	}

	if failed > 0 {
		parts = append(parts, statusStyleFailed.Render(fmt.Sprintf("%d failed", failed)))
	}

	if earlyExit > 0 {
		parts = append(parts, statusStyleEarlyExit.Render(fmt.Sprintf("%d skipped", earlyExit)))
	}

	if pending > 0 {
		parts = append(parts, statusStylePending.Render(fmt.Sprintf("%d pending", pending)))
	}

	return summaryStyle.Render(strings.Join(parts, "  "))
}

func statusIcon(s UnitStatus) string {
	switch s {
	case StatusPending:
		return "○"
	case StatusRunning:
		return "●"
	case StatusSucceeded:
		return "✓"
	case StatusFailed:
		return "✗"
	case StatusEarlyExit:
		return "⊘"
	default:
		return "?"
	}
}

func statusStyle(s UnitStatus) lipgloss.Style {
	switch s {
	case StatusPending:
		return statusStylePending
	case StatusRunning:
		return statusStyleRunning
	case StatusSucceeded:
		return statusStyleSucceeded
	case StatusFailed:
		return statusStyleFailed
	case StatusEarlyExit:
		return statusStyleEarlyExit
	default:
		return statusStylePending
	}
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return ""
	}

	seconds := d.Seconds()

	switch {
	case seconds < 60: //nolint:mnd
		return fmt.Sprintf("%.1fs", seconds)
	case seconds < 3600: //nolint:mnd
		m := int(seconds) / 60 //nolint:mnd
		s := int(seconds) % 60 //nolint:mnd
		return fmt.Sprintf("%dm%ds", m, s)
	default:
		h := int(seconds) / 3600        //nolint:mnd
		m := (int(seconds) % 3600) / 60 //nolint:mnd
		return fmt.Sprintf("%dh%dm", h, m)
	}
}
