package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LeperGnome/bt/internal/state"
	t "github.com/LeperGnome/bt/internal/tree"
	ui "github.com/LeperGnome/bt/internal/ui"
)

type model struct {
	windowHeight int
	windowWidth  int
	appState     state.State
	renderer     *ui.Renderer
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.windowWidth = msg.Width
	case tea.KeyMsg:
		return m, m.appState.ProcessKey(msg)
	}
	return m, nil
}
func (m model) View() string {
	return m.renderer.Render(m.appState, m.windowHeight, m.windowWidth)
}

func newModel(tree t.Tree, renderer ui.Renderer) model {
	s := state.NewState(&tree)
	return model{
		appState: s,
		renderer: &renderer,
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

	// TODO: sorting function as a flag?
	tree, err := t.InitTree(rootPath, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	renderer := ui.Renderer{EdgePadding: int(*paddingPtr)}
	m := newModel(tree, renderer)

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
