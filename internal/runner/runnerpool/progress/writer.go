package progress

import (
	"bytes"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// Writer implements io.Writer and sends output lines to a bubbletea Program
// as UnitOutputMsg messages. It buffers partial lines and only sends complete
// lines (ending with newline).
type Writer struct {
	unitPath string
	program  *tea.Program
	buffer   bytes.Buffer
	mu       sync.Mutex
}

// NewWriter creates a Writer that sends output for the given unit path
// to the provided bubbletea program.
func NewWriter(unitPath string, program *tea.Program) *Writer {
	return &Writer{
		unitPath: unitPath,
		program:  program,
	}
}

func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err := w.buffer.Write(p)
	if err != nil {
		return n, err
	}

	w.flushCompleteLines()

	return n, nil
}

// flushCompleteLines extracts complete lines from the buffer and sends
// each as a UnitOutputMsg to the bubbletea program.
func (w *Writer) flushCompleteLines() {
	for {
		line, err := w.buffer.ReadBytes('\n')
		if err != nil {
			// No complete line found; put back the partial data.
			w.buffer.Write(line)
			return
		}

		// Trim the trailing newline for display.
		text := string(bytes.TrimRight(line, "\n\r"))
		w.program.Send(UnitOutputMsg{
			Path: w.unitPath,
			Line: text,
		})
	}
}

// Flush sends any remaining buffered data as a final line.
func (w *Writer) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buffer.Len() > 0 {
		w.program.Send(UnitOutputMsg{
			Path: w.unitPath,
			Line: w.buffer.String(),
		})
		w.buffer.Reset()
	}
}
