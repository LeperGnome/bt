package tree

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	ErrNoSelectedChild = errors.New("no selected child or file is irregular")
	ErrNotDirectory    = errors.New("selected node is not a directory")
)

type Tree struct {
	Root        *Node
	CurrentDir  *Node
	Marked      *Node
	sortingFunc NodeSortingFunc
	watcher     *fsnotify.Watcher
	mu          sync.Mutex
}

func (t *Tree) GetSelectedChild() *Node {
	if len(t.CurrentDir.Children) > 0 {
		return t.CurrentDir.Children[t.CurrentDir.selectedChildIdx]
	}
	return nil
}

func (t *Tree) RefreshNodeParentByPath(path string) error {
	parentDir := filepath.Dir(path)
	cur := t.Root

	t.mu.Lock()
	defer t.mu.Unlock()

outer:
	for {
		if parentDir == cur.Path {
			return cur.readChildren(t.sortingFunc)
		}
		for _, ch := range cur.Children {
			if strings.HasPrefix(path, ch.Path) {
				cur = ch
				continue outer
			}
		}
		return nil
	}
}

func (t *Tree) RenameMarked(name string) error {
	if t.Marked == nil {
		return nil
	}
	newPath := filepath.Join(t.Marked.Parent.Path, name)
	err := os.Rename(t.Marked.Path, newPath)
	if err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}
	t.Marked = nil
	return nil
}

func (t *Tree) CreateFileInCurrent(name string) error {
	_, err := os.Create(filepath.Join(t.CurrentDir.Path, name))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	return nil
}

func (t *Tree) CreateDirectoryInCurrent(name string) error {
	err := os.Mkdir(filepath.Join(t.CurrentDir.Path, name), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func (t *Tree) ReadSelectedChildContent(buf []byte, limit int64) (int, error) {
	selectedNode := t.GetSelectedChild()
	if selectedNode == nil || !selectedNode.Info.Mode().IsRegular() {
		return 0, ErrNoSelectedChild
	}

	f, err := os.Open(selectedNode.Path)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	limitedReader := io.LimitReader(f, limit)
	n, err := limitedReader.Read(buf)
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}
	return n, nil
}

func (t *Tree) SelectNextChild() {
	if t.CurrentDir.selectedChildIdx < len(t.CurrentDir.Children)-1 {
		t.CurrentDir.selectedChildIdx++
	}
}

func (t *Tree) SelectPreviousChild() {
	if t.CurrentDir.selectedChildIdx > 0 {
		t.CurrentDir.selectedChildIdx--
	}
}

func (t *Tree) SetSelectedChildAsCurrent() error {
	selectedChild := t.GetSelectedChild()
	if selectedChild == nil {
		return nil
	}
	if !selectedChild.Info.IsDir() {
		return ErrNotDirectory
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if selectedChild.Children == nil {
		err := selectedChild.readChildren(t.sortingFunc)
		if err != nil {
			return fmt.Errorf("failed to read children: %w", err)
		}
		t.watcher.Add(selectedChild.Path)
	}
	t.CurrentDir = selectedChild
	return nil
}

func (t *Tree) SetParentAsCurrent() {
	if t.CurrentDir.Parent != nil {
		currentName := t.CurrentDir.Info.Name()
		newParentIdx := slices.IndexFunc(t.CurrentDir.Parent.Children, func(n *Node) bool { return n.Info.Name() == currentName })
		t.CurrentDir.Parent.selectedChildIdx = newParentIdx
		t.CurrentDir = t.CurrentDir.Parent
	}
}

func (t *Tree) MarkSelectedChild() bool {
	if selected := t.GetSelectedChild(); selected != nil {
		t.Marked = selected
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
	err := os.RemoveAll(t.Marked.Path)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	t.Marked = nil
	return nil
}

func (t *Tree) CopyMarkedToCurrentDir() error {
	if t.Marked == nil {
		return nil
	}
	targetDir := t.CurrentDir.Path
	targetFileName, err := generateNewFileName(t.Marked.Info.Name(), targetDir)
	if err != nil {
		return err
	}
	targetPath := filepath.Join(targetDir, targetFileName)

	err = os.Rename(t.Marked.Path, targetPath)
	if err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}
	t.Marked = nil
	return nil
}

func (t *Tree) MoveMarkedToCurrentDir() error {
	if t.Marked == nil {
		return nil
	}
	targetDir := t.CurrentDir.Path
	targetFileName, err := generateNewFileName(t.Marked.Info.Name(), targetDir)
	if err != nil {
		return err
	}
	targetPath := filepath.Join(targetDir, targetFileName)

	err = os.Rename(t.Marked.Path, targetPath)
	if err != nil {
		return fmt.Errorf("failed to move: %w", err)
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
			return fmt.Errorf("failed to read children: %w", err)
		}
		t.watcher.Add(selectedChild.Path)
	}
	return nil
}

func InitTree(dir string, sortingFunc NodeSortingFunc) (*Tree, <-chan NodeChange, error) {
	rootInfo, err := os.Lstat(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get info for %s: %w", dir, err)
	}
	if !rootInfo.IsDir() {
		return nil, nil, fmt.Errorf("%s is not a directory", dir)
	}
	if sortingFunc == nil {
		sortingFunc = defaultNodeSorting
	}

	root := &Node{
		Path:     dir,
		Info:     rootInfo,
		Parent:   nil,
		Children: []*Node{},
	}

	err = root.readChildren(sortingFunc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read children of root: %w", err)
	}
	if len(root.Children) == 0 {
		return nil, nil, fmt.Errorf("can't initialize on empty directory '%s'", dir)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}
	changeChan := runFSWatcher(watcher)
	err = watcher.Add(root.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add watcher: %w", err)
	}

	tree := &Tree{
		Root:        root,
		CurrentDir:  root,
		sortingFunc: sortingFunc,
		watcher:     watcher,
	}
	return tree, changeChan, nil
}

func generateNewFileName(fname, targetDir string) (string, error) {
	currentDirContent, err := os.ReadDir(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}
	for slices.ContainsFunc(currentDirContent, func(e fs.DirEntry) bool { return e.Name() == fname }) {
		fname = "copy_" + fname
	}
	return fname, nil
}
