package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LeperGnome/bt/internal/state"
	"github.com/LeperGnome/bt/internal/tree"
	ui "github.com/LeperGnome/bt/internal/ui"
)

type model struct {
	windowHeight int
	windowWidth  int

	appState *state.State
	renderer *ui.Renderer
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		listenFSEvents(m.appState.NodeChanges),
		listenPreviewReady(m.renderer.PreviewDoneChan),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.windowWidth = msg.Width
	case tea.KeyMsg:
		return m, m.appState.ProcessKey(msg)
	case tree.NodeChange:
		m.renderer.RemovePreviewCache(msg.Path)
		m.appState.ProcessNodeChange(msg)
		return m, listenFSEvents(m.appState.NodeChanges)
	case ui.Preview:
		m.renderer.SetPreviewCache(msg)
		return m, listenPreviewReady(m.renderer.PreviewDoneChan)
	}
	return m, nil
}
func (m model) View() string {
	return m.renderer.Render(m.appState, ui.Dimentions{Height: m.windowHeight, Width: m.windowWidth})
}

func newModel(root string, pad int, style ui.Stylesheet) (model, error) {
	s, err := state.InitState(root)
	if err != nil {
		return model{}, err
	}
	renderer := ui.NewRenderer(style, pad)
	return model{
		appState: s,
		renderer: renderer,
	}, nil
}

func listenFSEvents(eventsChan <-chan tree.NodeChange) tea.Cmd {
	return func() tea.Msg {
		return <-eventsChan
	}
}

func listenPreviewReady(previewChan <-chan ui.Preview) tea.Cmd {
	return func() tea.Msg {
		return <-previewChan
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

	m, err := newModel(rootPath, int(*paddingPtr), ui.DefaultStylesheet)
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
