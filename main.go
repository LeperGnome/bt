package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
)

type model struct {
	tree         *Tree
	renderer     *Renderer
	windowHeight int
	windowWidth  int
	moveBuff     *Node
	statusRow    string
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
		case "d":
			m.tree.MarkSelectedChild()
			m.statusRow = "moving " + m.tree.Marked.Path
		case "p":
			err := m.tree.MoveMarkedToCurrentDir()
			if err != nil {
				m.statusRow = err.Error()
			} else {
				m.statusRow = ""
			}
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
	selected := m.tree.GetSelectedChild()
	path := m.tree.CurrentDir.Path + "/..."
	changeTime := "--"
	size := int64(0)
	if selected != nil {
		path = selected.Path
		changeTime = selected.Info.ModTime().Format(time.RFC822)
		size = selected.Info.Size()
	}
	header := []string{
		color.GreenString("> " + path),
		color.MagentaString(fmt.Sprintf(
			"%v : %d bytes",
			changeTime,
			size,
		)),
		":" + m.statusRow,
	}
	renderedTree := m.renderer.Render(m.tree, m.windowHeight-len(header), m.windowWidth)

	return strings.Join(header, "\n") + "\n" + renderedTree
}

func newModel(tree Tree, renderer Renderer) model {
	return model{
		tree:     &tree,
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

	tree, err := InitTree(rootPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	renderer := Renderer{EdgePadding: int(*paddingPtr)}
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
