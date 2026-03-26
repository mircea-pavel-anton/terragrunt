package progress

import (
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	// maxOutputLines is the number of recent output lines kept per unit.
	maxOutputLines = 5
	// tickInterval is how often running timers are refreshed.
	tickInterval = 500 * time.Millisecond
)

// Terraform output patterns for detecting changes.
var (
	planChangesPattern  = regexp.MustCompile(`Plan: (\d+) to add, (\d+) to change, (\d+) to destroy`)
	applyChangesPattern = regexp.MustCompile(`Apply complete! Resources: (\d+) added, (\d+) changed, (\d+) destroyed`)
)

// unitState tracks the current state and output of a single unit.
type unitState struct {
	path          string
	displayPath   string
	status        UnitStatus
	startTime     time.Time
	endTime       time.Time
	lines         []string // ring buffer of last N output lines
	hasChanges    bool
	changeSummary string
}

// addLine appends a line to the ring buffer, evicting the oldest if full.
func (u *unitState) addLine(line string) {
	u.lines = append(u.lines, line)
	if len(u.lines) > maxOutputLines {
		u.lines = u.lines[len(u.lines)-maxOutputLines:]
	}

	// Detect terraform change patterns.
	if planChangesPattern.MatchString(line) {
		match := planChangesPattern.FindString(line)
		// Only flag as changes if not "0 to add, 0 to change, 0 to destroy"
		if !strings.Contains(match, "0 to add, 0 to change, 0 to destroy") {
			u.hasChanges = true
			u.changeSummary = match
		}
	}

	if applyChangesPattern.MatchString(line) {
		match := applyChangesPattern.FindString(line)
		if !strings.Contains(match, "0 added, 0 changed, 0 destroyed") {
			u.hasChanges = true
			u.changeSummary = match
		}
	}
}

// duration returns the elapsed time for this unit.
func (u *unitState) duration() time.Duration {
	if u.startTime.IsZero() {
		return 0
	}

	if !u.endTime.IsZero() {
		return u.endTime.Sub(u.startTime)
	}

	return time.Since(u.startTime)
}

// isExpanded returns true if this unit should show its output lines.
func (u *unitState) isExpanded() bool {
	switch u.status {
	case StatusRunning:
		return true
	case StatusFailed, StatusEarlyExit:
		return true
	case StatusSucceeded:
		return u.hasChanges
	default:
		return false
	}
}

// Model is the bubbletea model for the progress TUI.
type Model struct {
	units  map[string]*unitState
	order  []string // display order (unit paths)
	width  int
	height int
	done   bool
}

// NewModel creates a new progress Model with the given unit paths in display order.
func NewModel(unitPaths []string, displayPaths []string) Model {
	units := make(map[string]*unitState, len(unitPaths))

	for i, path := range unitPaths {
		dp := path
		if i < len(displayPaths) {
			dp = displayPaths[i]
		}

		units[path] = &unitState{
			path:        path,
			displayPath: dp,
			status:      StatusPending,
		}
	}

	return Model{
		units: units,
		order: unitPaths,
	}
}

// Init starts the tick timer for updating durations.
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// Update handles incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		if m.done {
			return m, nil
		}
		return m, tickCmd()

	case UnitStatusMsg:
		if u, ok := m.units[msg.Path]; ok {
			u.status = msg.Status
			switch msg.Status {
			case StatusRunning:
				u.startTime = time.Now()
			case StatusSucceeded, StatusFailed, StatusEarlyExit:
				u.endTime = time.Now()
			}
		}
		return m, nil

	case UnitOutputMsg:
		if u, ok := m.units[msg.Path]; ok {
			u.addLine(msg.Line)
		}
		return m, nil

	case DoneMsg:
		m.done = true
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
