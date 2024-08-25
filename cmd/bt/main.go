package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"

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
	selected := m.appState.Tree.GetSelectedChild()

	// NOTE: special case for empty dir
	path := m.appState.Tree.CurrentDir.Path + "/..."
	changeTime := "--"
	size := "0 B"

	if selected != nil {
		path = selected.Path
		changeTime = selected.Info.ModTime().Format(time.RFC822)
		size = formatSize(float64(selected.Info.Size()), 1024.0)
	}

	markedPath := ""
	if m.appState.Tree.Marked != nil {
		markedPath = m.appState.Tree.Marked.Path
	}

	operationBar := fmt.Sprintf(": %s", m.appState.OpBuf.Repr())
	if markedPath != "" {
		operationBar += fmt.Sprintf(" [%s]", markedPath)
	}

	if m.appState.OpBuf.IsInput() {
		s := lipgloss.
			NewStyle().
			Background(lipgloss.Color("#3C3C3C"))
		operationBar += fmt.Sprintf(" | %s |", s.Render(string(m.appState.InputBuf)))
	}

	// should probably render this somewhere else...
	header := []string{
		color.GreenString("> " + path),
		color.MagentaString(fmt.Sprintf(
			"%v : %s",
			changeTime,
			size,
		)),
		operationBar,
	}

	renderedTree := m.renderer.RenderTree(m.appState.Tree, m.windowHeight-len(header), m.windowWidth)

	return strings.Join(header, "\n") + "\n" + renderedTree
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

var sizes = [...]string{"b", "Kb", "Mb", "Gb", "Tb", "Pb", "Eb"}

func formatSize(s float64, base float64) string {
	unitsLimit := len(sizes)
	i := 0
	for s >= base && i < unitsLimit {
		s = s / base
		i++
	}
	f := "%.0f %s"
	if i > 1 {
		f = "%.2f %s"
	}
	return fmt.Sprintf(f, s, sizes[i])
}
