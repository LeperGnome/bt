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

type Operation int

const (
	Noop Operation = iota
	Move
	Copy
)

func (o Operation) Repr() string {
	return []string{"", "moving", "copying"}[o]
}

type model struct {
	tree         *Tree
	renderer     *Renderer
	windowHeight int
	windowWidth  int
	statusRow    string
	opBuf        Operation
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
		case "y":
			m.tree.MarkSelectedChild()
			m.opBuf = Copy
		case "d":
			m.tree.MarkSelectedChild()
			m.opBuf = Move
		case "p":
			switch m.opBuf {
			case Noop:
				break
			case Move:
				err := m.tree.MoveMarkedToCurrentDir()
				if err != nil {
					panic(err) // TODO
				}
			case Copy:
				err := m.tree.CopyMarkedToCurrentDir()
				if err != nil {
					panic(err) // TODO
				}
			}
			m.opBuf = Noop
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
	size := "0 B"
	if selected != nil {
		path = selected.Path
		changeTime = selected.Info.ModTime().Format(time.RFC822)
		size = formatSize(float64(selected.Info.Size()), 1024.0)
	}
	markedPath := ""
	if m.tree.Marked != nil {
		markedPath = m.tree.Marked.Path
	}
	header := []string{
		color.GreenString("> " + path),
		color.MagentaString(fmt.Sprintf(
			"%v : %s",
			changeTime,
			size,
		)),
		fmt.Sprintf(": %s %s", m.opBuf.Repr(), markedPath),
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
