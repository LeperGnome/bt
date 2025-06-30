package state

import (
	t "github.com/LeperGnome/bt/internal/tree"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"os/exec"
)

type Operation int

const (
	Noop Operation = iota
	Move
	Copy
	Delete
	Go
	Insert
	InsertFile
	InsertDir
	Rename
)

func (o Operation) Repr() string {
	return []string{
		"",
		"moving",
		"copying",
		"confirm removing (y/n) of",
		"g",
		"create new (f)ile/(d)irectory",
		"enter new file name:",
		"enter new directory name:",
		"renaming",
	}[o]
}
func (o Operation) IsInput() bool {
	switch o {
	case InsertDir, InsertFile, Rename:
		return true
	default:
		return false
	}
}

type State struct {
	Tree        *t.Tree
	OpBuf       Operation
	PrevOpBuf   Operation
	InputBuf    []rune
	ErrBuf      string
	NodeChanges <-chan t.NodeChange
	HelpToggle  bool
}

func InitState(root string) (*State, error) {
	tree, ncc, err := t.InitTree(root, nil)
	if err != nil {
		return nil, err
	}
	return &State{
		Tree:        tree,
		OpBuf:       Noop,
		InputBuf:    []rune{},
		NodeChanges: ncc,
	}, nil
}

func (s *State) ProcessNodeChange(nodeChange t.NodeChange) tea.Cmd {
	err := s.Tree.RefreshNodeParentByPath(nodeChange.Path)
	if err != nil {
		s.ErrBuf = err.Error()
	}
	s.Tree.RemoveNodeFromMarkByPath(nodeChange.Path)
	return nil
}

func (s *State) ProcessKey(msg tea.KeyMsg) tea.Cmd {
	switch s.OpBuf {
	case Noop:
		return s.processKeyDefault(msg)
	case Move:
		return s.processKeyMove(msg)
	case Delete:
		return s.processKeyDelete(msg)
	case Copy:
		return s.processKeyCopy(msg)
	case Go:
		return s.processKeyGo(msg)
	case Insert:
		return s.processKeyInsert(msg)
	case InsertFile:
		return s.processKeyInsertFile(msg)
	case InsertDir:
		return s.processKeyInsertDir(msg)
	case Rename:
		return s.processKeyRename(msg)
	default:
		return s.processKeyDefault(msg)
	}
}
func (s *State) processKeyRename(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		err := s.Tree.RenameMarked(string(s.InputBuf))
		if err != nil {
			s.ErrBuf = err.Error()
		}
		s.Tree.DropMark()
		s.OpBuf = Noop
		s.InputBuf = []rune{}
	default:
		return s.processKeyAnyInput(msg)
	}
	return nil
}
func (s *State) processKeyInsert(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "f":
		s.OpBuf = InsertFile
	case "d":
		s.OpBuf = InsertDir
	default:
		s.OpBuf = Noop
		return s.processKeyDefault(msg)
	}
	return nil
}
func (s *State) processKeyInsertFile(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		err := s.Tree.CreateFileInCurrent(string(s.InputBuf))
		if err != nil {
			s.ErrBuf = err.Error()
		}
		s.OpBuf = Noop
		s.InputBuf = []rune{}
	default:
		return s.processKeyAnyInput(msg)
	}
	return nil
}
func (s *State) processKeyInsertDir(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		err := s.Tree.CreateDirectoryInCurrent(string(s.InputBuf))
		if err != nil {
			s.ErrBuf = err.Error()
		}
		s.OpBuf = Noop
		s.InputBuf = []rune{}
	default:
		return s.processKeyAnyInput(msg)
	}
	return nil
}
func (s *State) processKeyAnyInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c", "esc":
		s.OpBuf = Noop
		s.PrevOpBuf = Noop
		s.InputBuf = []rune{}
		s.Tree.DropMark()
	case "backspace":
		if l := len(s.InputBuf); l > 0 {
			s.InputBuf = s.InputBuf[:l-1]
		}
	default:
		s.InputBuf = append(s.InputBuf, msg.Runes...)
	}
	return nil
}
func (s *State) processKeyGo(msg tea.KeyMsg) tea.Cmd {
	s.OpBuf = s.PrevOpBuf
	switch msg.String() {
	case "g":
		s.Tree.CurrentDir.SelectFirst()
	default:
		return s.processKeyDefault(msg)
	}
	return nil
}
func (s *State) processKeyDelete(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y":
		err := s.Tree.DeleteMarked()
		if err != nil {
			s.ErrBuf = err.Error()
		}
		s.OpBuf = Noop
	default:
		s.OpBuf = Noop
		s.Tree.DropMark()
		return s.processKeyDefault(msg)
	}
	return nil
}
func (s *State) processKeyMove(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "p":
		err := s.Tree.MoveMarkedToCurrentDir()
		if err != nil {
			s.ErrBuf = err.Error()
		}
		s.OpBuf = Noop
	default:
		return s.processKeyDefault(msg)
	}
	return nil
}
func (s *State) processKeyCopy(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "p":
		err := s.Tree.CopyMarkedToCurrentDir()
		if err != nil {
			s.ErrBuf = err.Error()
		}
		s.OpBuf = Noop
	default:
		return s.processKeyDefault(msg)
	}
	return nil
}
func (s *State) processKeyDefault(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		s.Tree.DropMark()
		s.OpBuf = Noop
		s.ErrBuf = ""
	case "ctrl+c", "q":
		return tea.Quit
	case "shift+tab":
		s.Tree.ToggleMarkSelectedChild()
		s.Tree.SelectPreviousChild()
	case "tab":
		s.Tree.ToggleMarkSelectedChild()
		s.Tree.SelectNextChild()
	case "j", "down":
		s.Tree.SelectNextChild()
	case "k", "up":
		s.Tree.SelectPreviousChild()
	case "l", "right":
		err := s.Tree.SetSelectedChildAsCurrent()
		if err != nil {
			s.ErrBuf = err.Error()
		}
	case "h", "left":
		s.Tree.SetParentAsCurrent()
	case "y":
		if len(s.Tree.Marked) != 0 {
			s.OpBuf = Copy
		} else if ok := s.Tree.MarkSelectedChild(); ok {
			s.OpBuf = Copy
		}
	case "d":
		if len(s.Tree.Marked) != 0 {
			s.OpBuf = Move
		} else if ok := s.Tree.MarkSelectedChild(); ok {
			s.OpBuf = Move
		}
	case "D":
		if len(s.Tree.Marked) != 0 {
			s.OpBuf = Delete
		} else if ok := s.Tree.MarkSelectedChild(); ok {
			s.OpBuf = Delete
		}
	case "g":
		s.PrevOpBuf = s.OpBuf
		s.OpBuf = Go
	case "G":
		s.Tree.CurrentDir.SelectLast()
	case "i":
		s.Tree.DropMark()
		s.OpBuf = Insert
	case "r":
		if ok := s.Tree.MarkSelectedChild(); ok && len(s.Tree.Marked) == 1 {
			s.InputBuf = []rune(s.Tree.Marked[0].Info.Name())
			s.OpBuf = Rename
		}
	case "e":
		child := s.Tree.GetSelectedChild()
		if child != nil && child.Info.Mode().IsRegular() {
			return openEditor(child.Path)
		}
	case "H":
		if err := s.Tree.ToggleHiddenInCurrentDirectory(); err != nil {
			s.ErrBuf = err.Error()
		}
	case "?":
		s.HelpToggle = !s.HelpToggle
	case "enter":
		child := s.Tree.GetSelectedChild()
		if child != nil && child.Info.Mode().IsRegular() {
			return xdgOpenFile(child.Path)
		} else {
			err := s.Tree.CollapseOrExpandSelected()
			if err != nil {
				s.ErrBuf = err.Error()
			}
		}
	}
	return nil
}

func openEditor(path string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return nil
	})
}

func xdgOpenFile(path string) tea.Cmd {
	c := exec.Command("xdg-open", path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return nil
	})
}
