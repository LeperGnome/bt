package tree

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type Tree struct {
	Root       *Node
	CurrentDir *Node
	Marked     []*Node

	sortingFunc NodeSortingFunc
	watcher     *fsnotify.Watcher
}

func (t *Tree) GetSelectedChild() *Node {
	if len(t.CurrentDir.Children) > 0 {
		return t.CurrentDir.Children[t.CurrentDir.selectedChildIdx]
	}
	return nil
}
func (t *Tree) ToggleHiddenInCurrentDirectory() error {
	t.CurrentDir.showHidden = !t.CurrentDir.showHidden
	return t.CurrentDir.readChildren(defaultNodeSorting)
}
func (t *Tree) RemoveNodeFromMarkByPath(path string) {
	t.Marked = slices.DeleteFunc(
		t.Marked,
		func(n *Node) bool { return n.Path == path },
	)
}
func (t *Tree) RefreshNodeParentByPath(path string) error {
	parentDir := filepath.Dir(path)
	cur := t.Root
outer:
	for {
		// Reading children when a parent node found.
		if parentDir == cur.Path {
			return cur.readChildren(t.sortingFunc)
		}
		// Going through directories towards `parentDir`.
		for _, ch := range cur.Children {
			if strings.HasPrefix(path, ch.Path) {
				cur = ch
				continue outer
			}
		}
		return nil
	}
}
func (t *Tree) RenameMarked(newName string) error {
	if len(t.Marked) != 1 {
		return nil
	}
	marked := t.Marked[0]

	if newName == "" {
		return fmt.Errorf("new name must not be empty")
	}
	targetPath := filepath.Join(marked.Parent.Path, newName)
	// NOTE: probably not the best way to check if file exists
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("file %s already exists", newName)
	} else if !os.IsNotExist(err) {
		return err
	}
	err := os.Rename(marked.Path, targetPath)
	if err != nil {
		return err
	}
	t.Marked = nil
	return nil
}
func (t *Tree) CreateFileInCurrent(name string) error {
	_, err := os.Create(filepath.Join(t.CurrentDir.Path, name))
	return err
}
func (t *Tree) CreateDirectoryInCurrent(name string) error {
	return os.Mkdir(filepath.Join(t.CurrentDir.Path, name), os.ModePerm)
}
func (t *Tree) SelectNextChild() {
	if t.CurrentDir.selectedChildIdx < len(t.CurrentDir.Children)-1 {
		t.CurrentDir.selectedChildIdx += 1
	}
}
func (t *Tree) SelectPreviousChild() {
	if t.CurrentDir.selectedChildIdx > 0 {
		t.CurrentDir.selectedChildIdx -= 1
	}
}
func (t *Tree) SetSelectedChildAsCurrent() error {
	selectedChild := t.GetSelectedChild()
	if selectedChild == nil {
		return nil
	}
	if !selectedChild.Info.IsDir() {
		return nil
	}
	if selectedChild.Children == nil {
		err := selectedChild.readChildren(t.sortingFunc)
		if err != nil {
			return err
		}
		t.watcher.Add(selectedChild.Path)
	}
	t.CurrentDir = selectedChild
	return nil
}
func (t *Tree) SetParentAsCurrent() {
	if t.CurrentDir.Parent != nil {
		currentName := t.CurrentDir.Info.Name()
		// setting parent current to match the directory, that we're leaving
		newParentIdx := slices.IndexFunc(t.CurrentDir.Parent.Children, func(n *Node) bool { return n.Info.Name() == currentName })
		t.CurrentDir.Parent.selectedChildIdx = newParentIdx

		t.CurrentDir = t.CurrentDir.Parent
	}
}
func (t *Tree) ToggleMarkSelectedChild() bool {
	if selected := t.GetSelectedChild(); selected != nil {
		if !slices.Contains(t.Marked, selected) {
			t.Marked = append(t.Marked, selected)
		} else {
			t.Marked = slices.DeleteFunc(t.Marked, func(n *Node) bool { return n == selected })
		}
		return true
	}
	return false
}
func (t *Tree) MarkSelectedChild() bool {
	if selected := t.GetSelectedChild(); selected != nil {
		if !slices.Contains(t.Marked, selected) {
			t.Marked = append(t.Marked, selected)
		}
		return true
	}
	return false
}
func (t *Tree) DropMark() {
	t.Marked = nil
}
func (t *Tree) DeleteMarked() error {
	if t.Marked == nil {
		return nil
	}
	for _, marked := range t.Marked {
		cmd := exec.Command("rm", "-r", marked.Path)
		err := cmd.Run()
		if err != nil {
			return err // todo: this is not the same error...?
		}
	}
	t.Marked = nil
	return nil
}
func (t *Tree) CopyMarkedToCurrentDir() error {
	if t.Marked == nil {
		return nil
	}
	targetDir := t.CurrentDir.Path
	for _, marked := range t.Marked {
		targetFileName, err := generateNewFileName(marked.Info.Name(), targetDir)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, targetFileName)

		cmd := exec.Command("cp", "-r", marked.Path, targetPath)
		err = cmd.Run()
		if err != nil {
			return err // todo: this is not the same error...?
		}
	}
	t.Marked = nil
	return nil
}
func (t *Tree) MoveMarkedToCurrentDir() error {
	if t.Marked == nil {
		return nil
	}
	targetDir := t.CurrentDir.Path
	for _, marked := range t.Marked {
		targetFileName, err := generateNewFileName(marked.Info.Name(), targetDir)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, targetFileName)

		cmd := exec.Command("mv", "-n", marked.Path, targetPath)
		err = cmd.Run()
		if err != nil {
			return err // todo: this is not the same error...?
		}
	}
	t.Marked = nil
	return nil
}
func (t *Tree) CollapseOrExpandSelected() error {
	selectedChild := t.GetSelectedChild()
	if selectedChild == nil {
		return nil
	}
	if selectedChild.Children != nil {
		selectedChild.orphanChildren()
		t.watcher.Remove(selectedChild.Path)
	} else {
		err := selectedChild.readChildren(t.sortingFunc)
		if err != nil {
			return err
		}
		t.watcher.Add(selectedChild.Path)
	}
	return nil
}

func InitTree(dir string, sortingFunc NodeSortingFunc) (*Tree, <-chan NodeChange, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, nil, err
	}

	rootInfo, err := os.Lstat(absDir)
	if err != nil {
		return nil, nil, err
	}
	if !rootInfo.IsDir() {
		return nil, nil, fmt.Errorf("%s is not a directory", absDir)
	}
	if sortingFunc == nil {
		sortingFunc = defaultNodeSorting
	}

	root := NewNode(absDir, rootInfo, nil)

	err = root.readChildren(sortingFunc)
	if err != nil {
		return nil, nil, err
	}
	if len(root.Children) == 0 {
		return nil, nil, fmt.Errorf("Can't initialize on empty directory '%s'", absDir)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}
	changeChan := runFSWatcher(watcher)
	err = watcher.Add(root.Path)
	if err != nil {
		return nil, nil, err
	}

	tree := &Tree{
		Root:        root,
		CurrentDir:  root,
		sortingFunc: sortingFunc,
		watcher:     watcher,
	}
	return tree, changeChan, nil
}

// Checks if fname already exists in targetDir.
// Adds "copy_" prefix (multiple times), until new file name becomes unique in derecotry.
func generateNewFileName(fname, targetDir string) (string, error) {
	currentDirContent, err := os.ReadDir(targetDir)
	if err != nil {
		return "", err
	}
	for slices.ContainsFunc(currentDirContent, func(e fs.DirEntry) bool { return e.Name() == fname }) {
		fname = "copy_" + fname
	}
	return fname, nil
}
