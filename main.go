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
	tree          Tree
	fixEdgeOffset int
	winHeight     int
	offsetMem     *int
}

func (m model) Init() tea.Cmd {
	return nil
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.winHeight = msg.Height

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
			m.winHeight,
		)),
	}
	// max height for 'tree' section wihtout header
	maxHeight := m.winHeight - len(lines)

	// rendering tree
	rendered, selectedRow := Render(&m.tree)
	totalTreeLines := len(rendered)

	// determining offset and limit based on selected row
	offset := *m.offsetMem
	limit := totalTreeLines
	if maxHeight > 0 {
		// cursor is out for 'top' boundary
		if selectedRow+1 > maxHeight+offset-m.fixEdgeOffset {
			offset = min(selectedRow+1-maxHeight+m.fixEdgeOffset, totalTreeLines-maxHeight)
		}
		// cursor is out for 'bottom' boundary
		if selectedRow < m.fixEdgeOffset+offset {
			offset = max(selectedRow-m.fixEdgeOffset, 0)
		}
		*m.offsetMem = offset
		limit = min(maxHeight+offset, totalTreeLines)
	}

	lines = append(lines, rendered[offset:limit]...)
	return strings.Join(lines, "\n")
}

func newModel(tree Tree, edgeOffset int) model {
	offsetMem := 0
	return model{
		tree:          tree,
		offsetMem:     &offsetMem,
		fixEdgeOffset: edgeOffset,
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
