// Package progress provides a Docker BuildKit-style TUI for displaying
// parallel unit execution progress in terragrunt run --all commands.
package progress

import "time"

// UnitStatusMsg is sent when a unit's execution status changes.
type UnitStatusMsg struct {
	Path   string
	Status UnitStatus
}

// UnitOutputMsg is sent when a unit produces a line of output.
type UnitOutputMsg struct {
	Path string
	Line string
}

// DoneMsg signals that all units have finished execution.
type DoneMsg struct{}

// TickMsg is sent periodically to update running duration timers.
type TickMsg time.Time

// UnitStatus represents the execution state of a unit.
type UnitStatus int

const (
	StatusPending UnitStatus = iota
	StatusRunning
	StatusSucceeded
	StatusFailed
	StatusEarlyExit
)

// String returns a human-readable label for the status.
func (s UnitStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusSucceeded:
		return "succeeded"
	case StatusFailed:
		return "failed"
	case StatusEarlyExit:
		return "early exit"
	default:
		return "unknown"
	}
}
