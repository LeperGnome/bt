package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LeperGnome/bt/internal/state"
	ui "github.com/LeperGnome/bt/internal/ui"
)

type model struct {
	windowHeight int
	windowWidth  int
	appState     *state.State
	renderer     *ui.Renderer
	eventsChan   <-chan state.NodeChange
}

func (m model) Init() tea.Cmd {
	return listenFSEvents(m.eventsChan)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.windowWidth = msg.Width
	case tea.KeyMsg:
		return m, m.appState.ProcessKey(msg)
	case state.NodeChange:
		m.appState.ProcessNodeChange(msg)
		return m, listenFSEvents(m.eventsChan)
	}
	return m, nil
}
func (m model) View() string {
	return m.renderer.Render(m.appState, m.windowHeight, m.windowWidth)
}

func newModel(root string, pad int) (model, error) {
	s, err := state.InitState(root)
	if err != nil {
		return model{}, err
	}
	renderer := &ui.Renderer{EdgePadding: pad}
	return model{
		appState:   s,
		renderer:   renderer,
		eventsChan: state.InitFakeFSWatcher(),
	}, nil
}

func listenFSEvents(eventsChan <-chan state.NodeChange) tea.Cmd {
	return func() tea.Msg {
		return <-eventsChan
	}
}

func main() {
	paddingPtr := flag.Uint("pad", 5, "Edge padding for top and bottom")
	inlinePtr := flag.Bool("i", false, "In-place render (without alternate screen)")
	flag.Parse()
	rootPath := flag.Arg(0)
	if rootPath == "" {
		rootPath = "."
	}

	m, err := newModel(rootPath, int(*paddingPtr))
	if err != nil {
		fmt.Printf("Error on init: %v", err)
		os.Exit(1)
	}

	opts := []tea.ProgramOption{}
	if !*inlinePtr {
		opts = append(opts, tea.WithAltScreen())
	}

	p := tea.NewProgram(m, opts...)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
