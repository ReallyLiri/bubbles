package viewport

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MODEL

type Model struct {
	Err    error
	Width  int
	Height int

	// YOffset is the vertical scroll position.
	YOffset int

	// YPosition is the position of the viewport in relation to the terminal
	// window. It's used in high performance rendering.
	YPosition int

	// HighPerformanceRendering bypasses the normal Bubble Tea renderer to
	// provide higher performance rendering. Most of the time the normal Bubble
	// Tea rendering methods will suffice, but if you're passing content with
	// a lot of ANSI escape codes you may see improved rendering in certain
	// terminals with this enabled.
	//
	// This should only be used in program occupying the entire terminal width,
	// usually via the alternate screen buffer.
	HighPerformanceRendering bool

	lines []string
}

func NewModel(width, height int) Model {
	return Model{
		Width:  width,
		Height: height,
	}
}

func (m Model) SetSize(yPos int, width, height int) {
	m.YPosition = yPos
	m.Width = width
	m.Height = height
}

// AtTop returns whether or not the viewport is in the very top position.
func (m Model) AtTop() bool {
	return m.YOffset <= 0
}

// AtBottom returns whether or not the viewport is at the very botom position.
func (m Model) AtBottom() bool {
	return m.YOffset >= len(m.lines)-m.Height-1
}

// Scrollpercent returns the amount scrolled as a float between 0 and 1.
func (m Model) ScrollPercent() float64 {
	if m.Height >= len(m.lines) {
		return 1.0
	}
	y := float64(m.YOffset)
	h := float64(m.Height)
	t := float64(len(m.lines))
	return y / (t - h)
}

// SetContent set the pager's text content. For high performance rendering the
// Sync command should also be called.
func (m *Model) SetContent(s string) {
	s = strings.Replace(s, "\r\n", "\n", -1) // normalize line endings
	m.lines = strings.Split(s, "\n")
}

// ViewDown moves the view down by the number of lines in the viewport.
// Basically, "page down".
func (m *Model) ViewDown() {
	if m.AtBottom() {
		return
	}

	m.YOffset = min(
		m.YOffset+m.Height,    // target
		len(m.lines)-m.Height, // fallback
	)
}

// ViewUp moves the view up by one height of the viewport. Basically, "page up".
func (m *Model) ViewUp() {
	if m.AtTop() {
		return
	}

	m.YOffset = max(
		m.YOffset-m.Height, // target
		0,                  // fallback
	)
}

// HalfViewDown moves the view down by half the height of the viewport.
func (m *Model) HalfViewDown() {
	if m.AtBottom() {
		return
	}

	m.YOffset = min(
		m.YOffset+m.Height/2,  // target
		len(m.lines)-m.Height, // fallback
	)
}

// HalfViewUp moves the view up by half the height of the viewport.
func (m *Model) HalfViewUp() {
	if m.AtTop() {
		return
	}

	m.YOffset = max(
		m.YOffset-m.Height/2, // target
		0,                    // fallback
	)
}

// LineDown moves the view up by the given number of lines.
func (m *Model) LineDown(n int) {
	if m.AtBottom() || n == 0 {
		return
	}

	m.YOffset = min(
		m.YOffset+n,           // target
		len(m.lines)-m.Height, // fallback
	)
}

// LineUp moves the view down by the given number of lines.
func (m *Model) LineUp(n int) {
	if m.AtTop() || n == 0 {
		return
	}

	m.YOffset = max(m.YOffset-n, 0)
}

// COMMANDS

func Sync(m Model) tea.Cmd {
	top := max(m.YOffset, 0)
	bottom := min(m.YOffset+m.Height, len(m.lines)-1)

	return tea.SyncScrollArea(
		m.lines[top:bottom],
		m.YPosition,
		m.YPosition+m.Height,
	)
}

func ViewDown(m Model) tea.Cmd {
	if m.AtBottom() {
		return nil
	}

	top := max(m.YOffset+m.Height, 0)
	bottom := min(top+m.Height, len(m.lines)-1)

	return tea.ScrollDown(
		m.lines[top:bottom],
		m.YPosition,
		m.YPosition+m.Height,
	)
}

func ViewUp(m Model) tea.Cmd {
	if m.AtTop() {
		return nil
	}

	top := max(m.YOffset-m.Height, 0)
	bottom := min(m.YOffset, len(m.lines)-1)

	return tea.ScrollUp(
		m.lines[top:bottom],
		m.YPosition,
		m.YPosition+m.Height,
	)
}

func HalfViewDown(m Model) tea.Cmd {
	if m.AtBottom() {
		return nil
	}

	top := max(m.YOffset+m.Height/2, 0)
	bottom := min(top+m.Height, len(m.lines)-1)

	return tea.ScrollDown(
		m.lines[top:bottom],
		m.YPosition,
		m.YPosition+m.Height,
	)
}

func HalfViewUp(m Model) tea.Cmd {
	if m.AtTop() {
		return nil
	}

	top := max(m.YOffset-m.Height/2, 0)
	bottom := clamp(m.YOffset, top, len(m.lines)-1)

	return tea.ScrollUp(
		m.lines[top:bottom],
		m.YPosition,
		m.YPosition+m.Height,
	)
}

func LineDown(m Model, n int) tea.Cmd {
	if m.AtBottom() || n == 0 {
		return nil
	}

	bottom := min(m.YOffset+m.Height+n, len(m.lines)-1)
	top := max(bottom-n, 0)

	return tea.ScrollDown(
		m.lines[top:bottom],
		m.YPosition,
		m.YPosition+m.Height,
	)
}

func LineUp(m Model, n int) tea.Cmd {
	if m.AtTop() || n == 0 {
		return nil
	}

	top := max(m.YOffset-n, 0)
	bottom := min(top+n, len(m.lines)-1)

	return tea.ScrollUp(
		m.lines[top:bottom],
		m.YPosition,
		m.YPosition+m.Height,
	)
}

// UPDATE

// Update runs the update loop with default keybindings similar to popular
// pagers. To define your own keybindings use the methods on Model (i.e.
// Model.LineDown()) and define your own update function.
func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		// Down one page
		case "pgdown":
			fallthrough
		case " ": // spacebar
			fallthrough
		case "f":
			if m.HighPerformanceRendering {
				cmd = ViewDown(m)
			}
			m.ViewDown()

		// Up one page
		case "pgup":
			fallthrough
		case "b":
			if m.HighPerformanceRendering {
				cmd = ViewUp(m)
			}
			m.ViewUp()

		// Down half page
		case "d":
			if m.HighPerformanceRendering {
				cmd = HalfViewDown(m)
			}
			m.HalfViewDown()

		// Up half page
		case "u":
			if m.HighPerformanceRendering {
				cmd = HalfViewUp(m)
			}
			m.HalfViewUp()

		// Down one line
		case "down":
			fallthrough
		case "j":
			if m.HighPerformanceRendering {
				cmd = LineDown(m, 1)
			}
			m.LineDown(1)

		// Up one line
		case "up":
			fallthrough
		case "k":
			if m.HighPerformanceRendering {
				cmd = LineUp(m, 1)
			}
			m.LineUp(1)
		}
	}

	return m, cmd
}

// VIEW

// View renders the viewport into a string.
func View(m Model) string {

	if m.HighPerformanceRendering {
		// Just send newlines  since we're doing to be rendering the actual
		// content seprately. We do need to send something so that the Bubble
		// Tea standard renderer can push everything else down.
		return strings.Repeat("\n", m.Height-1)
	}

	if m.Err != nil {
		return m.Err.Error()
	}

	var lines []string

	if len(m.lines) > 0 {
		top := max(0, m.YOffset)
		bottom := min(len(m.lines), m.YOffset+m.Height)
		lines = m.lines[top:bottom]
	}

	// Fill empty space with newlines
	extraLines := ""
	if len(lines) < m.Height {
		extraLines = strings.Repeat("\n", m.Height-len(lines))
	}

	return strings.Join(lines, "\n") + extraLines
}

// ETC

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(val, low, high int) int {
	return max(low, min(high, val))
}
