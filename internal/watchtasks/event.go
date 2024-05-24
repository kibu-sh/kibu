package watchtasks

type EventType int

func (e EventType) String() string {
	switch e {
	case EventTypeNone:
		return "none"
	case EventTypeRebuilding:
		return "rebuilding"
	case EventTypeDebuggerRestarted:
		return "debugger restarted"
	case EventTypeDebuggerListening:
		return "debugger listening"
	case EventTypeError:
		return "error"
	default:
		return "unknown"
	}
}

const (
	EventTypeNone EventType = iota
	EventTypeRebuilding
	EventTypeDebuggerRestarted
	EventTypeDebuggerListening
	EventTypeError
)

type Event struct {
	Type EventType
	Err  error
}
