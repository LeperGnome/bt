package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	root    *Node
	currentDir *Node
}

func (a App) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "j", "down":
			if a.currentDir.Selected < len(a.currentDir.Children)-1 {
				a.currentDir.Selected += 1
				a.currentDir.smem += 1
			}
		case "k", "up":
			if a.currentDir.Selected > 0 {
				a.currentDir.Selected -= 1
				a.currentDir.smem -= 1
			}
		case "l", "right":
            next := &a.currentDir.Children[a.currentDir.Selected]
			if next.Children == nil {
				err := next.ReadChildren()
				if err != nil {
					panic(err) // TODO
				}
			}
			if len(next.Children) > 0 {
				a.currentDir = next
                a.currentDir.Selected = a.currentDir.smem
			}
		case "h", "left":
			if a.currentDir.Parent != nil {
				a.currentDir.Selected = NotSelected
				a.currentDir = a.currentDir.Parent
			}
		case "enter":
            selectedNode := &a.currentDir.Children[a.currentDir.Selected]
            if selectedNode.Children != nil {
                selectedNode.OrphanChildren()
            } else {
                err := selectedNode.ReadChildren()
                if err != nil {
                    panic(err) // TODO
                }
            }
		}
	}
	return a, nil
}
func (a App) View() string {
	// The header
	s := "> " + a.root.Path + "\n"
    s += fmt.Sprintf("current: '%s' selected child = %d\n", a.currentDir.Info.Name(), a.currentDir.Selected)
	s += a.root.View()

	return s
}

func main() {
	flag.Parse()
	rootPath := flag.Arg(0)
	if rootPath == "" {
		rootPath = "."
	}

	root, err := InitRoot(rootPath)
	if err != nil {
		panic(err)
	}

	app := App{
		root:    root,
		currentDir: root,
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
