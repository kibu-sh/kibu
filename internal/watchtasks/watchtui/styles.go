package watchtui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/discernhq/devx/internal/watchtasks"
)

var baseStatusStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#4D4D4D")).
	Bold(true).
	Align(lipgloss.Center).
	Padding(1).
	Width(30)

var statusInfoStyle = baseStatusStyle.Copy().
	Background(lipgloss.Color("#125471"))

var statusErrorStyle = baseStatusStyle.Copy().
	Background(lipgloss.Color("#7E3531"))

var statusSuccessStyle = baseStatusStyle.Copy().
	Background(lipgloss.Color("#335934"))

func chooseMessageForEventType(t watchtasks.EventType) string {
	switch t {
	case watchtasks.EventTypeNone:
		return "initializing"
	case watchtasks.EventTypeRebuilding:
		return "rebuilding"
	case watchtasks.EventTypeError:
		return "error"
	case watchtasks.EventTypeDebuggerListening,
		watchtasks.EventTypeDebuggerRestarted:
		return "ready, waiting for changes..."
	default:
		return "unknown"
	}
}

func chooseStyleForEventType(t watchtasks.EventType) lipgloss.Style {
	switch t {
	case watchtasks.EventTypeNone:
		return baseStatusStyle
	case watchtasks.EventTypeRebuilding:
		return statusInfoStyle
	case watchtasks.EventTypeError:
		return statusErrorStyle
	case watchtasks.EventTypeDebuggerListening,
		watchtasks.EventTypeDebuggerRestarted:
		return statusSuccessStyle
	default:
		return baseStatusStyle
	}
}

func renderStatusStyle(t watchtasks.EventType) string {
	return chooseStyleForEventType(t).Render(chooseMessageForEventType(t))
}
