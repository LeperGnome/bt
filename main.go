package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
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
			m.moveBuff = m.tree.GetSelectedChild()
			m.statusRow = "moving " + m.moveBuff.Path

		case "p":
			if m.moveBuff != nil {
				target := m.tree.CurrentDir.Path
				cmd := exec.Command("mv", m.moveBuff.Path, target)
				err := cmd.Run()
				if err != nil {
					m.statusRow = "error moving file - " + err.Error()
				}
				err = m.tree.CurrentDir.ReadChildren()
				if err != nil {
					panic(err) // TODO
				}
				err = m.moveBuff.Parent.ReadChildren()
				if err != nil {
					panic(err) // TODO
				}
				m.moveBuff = nil
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
	header := []string{
		color.GreenString("> " + selected.Path),
		color.MagentaString(fmt.Sprintf(
			"%v : %d bytes : current = %s : selected idx = %d",
			selected.Info.ModTime().Format(time.RFC822),
			selected.Info.Size(),
			m.tree.CurrentDir.Path, m.tree.CurrentDir.SelectedChildIdx,
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
