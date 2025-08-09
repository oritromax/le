package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdp/qrterminal/v3"
	"go.sakib.dev/le/server"
)

type model struct {
	srvr *server.Server
}

func newModel(srvr *server.Server) model {
	return model{
		srvr: srvr,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	case string:
		if msg == "update" {
			// Handle update messages, e.g., refresh the view
			return m, nil
		}
	default:
		// Handle other messages if necessary
	}

	return m, nil
}

func (m model) View() string {
	state := m.srvr.GetState()
	if state.Addr == nil {
		// return a loading indicator
		return "Loading server address...\nPress Ctrl+C or 'q' to quit.\n"
	}

	stringWriter := &strings.Builder{}

	qrterminal.GenerateWithConfig(*state.Addr, qrterminal.Config{
		Level:      qrterminal.L,
		Writer:     stringWriter,
		HalfBlocks: true,
		BlackChar:  qrterminal.BLACK_BLACK,
	})

	connCount := len(state.Conns)
	return fmt.Sprintf("Server running at: %s\nNumber of connections: %d\n%s\nPress Ctrl+C or 'q' to quit.\n", *state.Addr, connCount, stringWriter.String())
}

func Start(srvr *server.Server, ch <-chan server.ServerEventName) error {
	p := tea.NewProgram(newModel(srvr), tea.WithAltScreen())

	go func() {
		for range ch {
			p.Send(tea.Msg("update"))
		}
	}()

	// Save original stdout
	old := os.Stdout

	// Redirect stdout to /dev/null
	devNull, _ := os.Open(os.DevNull)
	os.Stdout = devNull

	if _, err := p.Run(); err != nil {
		return err
	}

	os.Stdout = old // Restore original stdout
	return nil
}
