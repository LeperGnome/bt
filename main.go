package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
)

type model struct {
	tree         *Tree
	renderer     *Renderer
	windowHeight int
}

func (m model) Init() tea.Cmd {
	return nil
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "j", "down":
			m.tree.SelectNextChild()
		case "k", "up":
			m.tree.SelectPreviousChild()
		case "l", "right":
			err := m.tree.SetSelectedChildAsCurrent()
			if err != nil {
				panic(err) // TODO
			}
		case "h", "left":
			m.tree.SetParentAsCurrent()
		case "enter":
			err := m.tree.CollapseOrExpandSelected()
			if err != nil {
				panic(err) // TODO
			}
		}
	}
	return m, nil
}
func (m model) View() string {
	lines := []string{
		color.GreenString("> " + m.tree.GetSelectedChild().Path),
		color.MagentaString(fmt.Sprintf(
			"current = '%s'; selected child idx = %d; winH = %d",
			m.tree.CurrentDir.Info.Name(),
			m.tree.CurrentDir.Selected,
			m.windowHeight,
		)),
	}
	rendered := m.renderer.Render(m.tree, m.windowHeight-len(lines))

	lines = append(lines, rendered...)
	return strings.Join(lines, "\n")
}

func newModel(tree Tree, edgePadding int) model {
	return model{
		tree:     &tree,
		renderer: &Renderer{EdgePadding: edgePadding},
	}
}

func main() {
	flag.Parse()
	rootPath := flag.Arg(0)
	if rootPath == "" {
		rootPath = "."
	}

	tree, err := InitTree(rootPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	m := newModel(tree, 5)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
