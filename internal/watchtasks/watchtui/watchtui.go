package watchtui

import (
	"context"
	"fmt"
	"github.com/discernhq/devx/internal/watchtasks"
	"github.com/discernhq/devx/pkg/messaging"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
)

type rootModel struct {
	stopwatch     stopwatch.Model
	keymap        keymap
	help          help.Model
	quitting      bool
	ctx           context.Context
	lastEventType watchtasks.EventType
	events        messaging.Stream[watchtasks.Event]
	width         int
	height        int
}

type keymap struct {
	start   key.Binding
	stop    key.Binding
	rebuild key.Binding
	quit    key.Binding
}

func watchEventStream(ctx context.Context, events messaging.Stream[watchtasks.Event]) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			events.Unsubscribe()
			return tea.Quit
		case e := <-events.Channel():
			return e
		}
	}
}

func (m rootModel) Init() tea.Cmd {
	return tea.Batch(
		m.stopwatch.Init(),
		watchEventStream(m.ctx, m.events),
	)
}

func (m rootModel) View() string {
	// Note: you could further customize the time output by getting the
	// duration from m.stopwatch.Elapsed(), which returns a time.Duration, and
	// skip m.stopwatch.View() altogether.
	b := strings.Builder{}

	if !m.quitting {
		b.WriteString(fmt.Sprintf("%s %s\n",
			renderStatusStyle(m.lastEventType),
			m.stopwatch.View(),
		))
		b.WriteString(m.helpView())
	}

	return b.String()
}

func (m rootModel) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.start,
		m.keymap.stop,
		m.keymap.rebuild,
		m.keymap.quit,
	})
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case watchtasks.Event:
		m.lastEventType = msg.Type
		batch := []tea.Cmd{
			watchEventStream(m.ctx, m.events),
		}

		switch m.lastEventType {
		case watchtasks.EventTypeRebuilding:
			batch = append(batch,
				m.stopwatch.Reset(),
				m.stopwatch.Start())
		case watchtasks.EventTypeDebuggerListening,
			watchtasks.EventTypeDebuggerRestarted:
			batch = append(batch, m.stopwatch.Stop())
		default:
			if !m.stopwatch.Running() {
				batch = append(batch, m.stopwatch.Start())
			}
		}
		return m, tea.Batch(batch...)
	}
	var cmd tea.Cmd
	m.stopwatch, cmd = m.stopwatch.Update(msg)
	return m, cmd
}

func Start(ctx context.Context, events messaging.Stream[watchtasks.Event]) error {
	m := rootModel{
		ctx:       ctx,
		events:    events,
		stopwatch: stopwatch.NewWithInterval(time.Millisecond),
		keymap: keymap{
			start: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "start"),
			),
			stop: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "stop"),
			),
			rebuild: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "rebuild"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "q"),
				key.WithHelp("q", "quit"),
			),
		},
		help: help.New(),
	}

	m.keymap.start.SetEnabled(false)
	p := tea.NewProgram(m,
		tea.WithContext(ctx),
		tea.WithAltScreen(),
		tea.WithoutSignalHandler(),
	)

	_, err := p.Run()
	return err
}
