package input

import (
	"time"

	"github.com/charmbracelet/tea"
)

type Model struct {
	Prompt       string
	Value        string
	Cursor       string
	HiddenCursor string
	BlinkSpeed   time.Duration

	blink bool
	pos   int
}

type CursorBlinkMsg struct{}

func DefaultModel() Model {
	return Model{
		Prompt:     "> ",
		Value:      "",
		BlinkSpeed: time.Millisecond * 600,

		blink: false,
		pos:   0,
	}
}

func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace:
			fallthrough
		case tea.KeyDelete:
			if len(m.Value) > 0 {
				m.Value = m.Value[:m.pos-1] + m.Value[m.pos:]
				m.pos--
			}
			return m, nil
		case tea.KeyLeft:
			if m.pos > 0 {
				m.pos--
			}
			return m, nil
		case tea.KeyRight:
			if m.pos < len(m.Value) {
				m.pos++
			}
			return m, nil
		case tea.KeyCtrlA: // ^A, beginning
			m.pos = 0
			return m, nil
		case tea.KeyCtrlD: // ^D, delete char under cursor
			if len(m.Value) > 0 && m.pos < len(m.Value) {
				m.Value = m.Value[:m.pos] + m.Value[m.pos+1:]
			}
			return m, nil
		case tea.KeyCtrlE: // ^E, end
			m.pos = len(m.Value) - 1
			return m, nil
		case tea.KeyCtrlK: // ^K, kill text after cursor
			m.Value = m.Value[:m.pos]
			m.pos = len(m.Value)
			return m, nil
		case tea.KeyRune:
			m.Value = m.Value[:m.pos] + msg.String() + m.Value[m.pos:]
			m.pos++
			return m, nil
		default:
			return m, nil
		}

	case CursorBlinkMsg:
		m.blink = !m.blink
		return m, nil

	default:
		return m, nil
	}
}

func View(model tea.Model) string {
	m, _ := model.(Model)
	v := m.Value[:m.pos]
	if m.pos < len(m.Value) {
		v += cursor(string(m.Value[m.pos]), m.blink)
		v += m.Value[m.pos+1:]
	} else {
		v += cursor(" ", m.blink)
	}
	return m.Prompt + v
}

// Style the cursor
func cursor(s string, blink bool) string {
	if blink {
		return s
	}
	return tea.Invert(s)
}

// Subscription
func Blink(model tea.Model) tea.Msg {
	m, ok := model.(Model)
	if !ok {
		return tea.NewErrMsg("could not assert given model to the model we expected; make sure you're passing as input model")
	}
	time.Sleep(m.BlinkSpeed)
	return CursorBlinkMsg{}
}
