package tree

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type Tree struct {
	Root        *Node
	CurrentDir  *Node
	Marked      *Node
	sortingFunc NodeSortingFunc
	watcher     *fsnotify.Watcher
	searchIndex map[string]*Node
}

func (t *Tree) GetSelectedChild() *Node {
	if len(t.CurrentDir.Children) > 0 {
		return t.CurrentDir.Children[t.CurrentDir.selectedChildIdx]
	}
	return nil
}
func (t *Tree) SearchNodes(name string) []*Node {
	var res []*Node
	for n, ptr := range t.searchIndex {
		if strings.Contains(n, name) {
			res = append(res, ptr)
		}
	}
	return res
}
func (t *Tree) RefreshNodeParentByPath(path string) error {
	// I'm assuming, that all paths are relative to my tree root
	parentDir := filepath.Dir(path)
	cur := t.Root
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
	err := os.Rename(t.Marked.Path, filepath.Join(t.Marked.Parent.Path, name))
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
func (t *Tree) ReadSelectedChildContent(buf []byte, limit int64) (int, error) {
	selectedNode := t.GetSelectedChild()
	if selectedNode == nil || !selectedNode.Info.Mode().IsRegular() {
		return 0, fmt.Errorf("file not selected or is irregular")
	}
	f, err := os.Open(selectedNode.Path)
	defer f.Close()
	if err != nil {
		return 0, err
	}
	limitedReader := io.LimitReader(f, limit)
	n, err := limitedReader.Read(buf)
	if err != nil {
		return 0, err
	}
	return n, nil
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
	// In an empty directory.
	if selectedChild == nil {
		return nil
	}
	// Directory is currently collapsed.
	if selectedChild.Children == nil {
		err := t.openDir(selectedChild)
		if err != nil {
			return err
		}
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
	cmd := exec.Command("rm", "-r", t.Marked.Path)
	err := cmd.Run()
	if err != nil {
		return err // todo: this is not the same error...?
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

	cmd := exec.Command("cp", "-r", t.Marked.Path, targetPath)
	err = cmd.Run()
	if err != nil {
		return err // todo: this is not the same error...?
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

	cmd := exec.Command("mv", t.Marked.Path, targetPath)
	err = cmd.Run()
	if err != nil {
		return err // todo: this is not the same error...?
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
		t.closeDir(selectedChild)
	} else {
		return t.openDir(selectedChild)
	}
	return nil
}
func (t *Tree) closeDir(dir *Node) {
	dir.orphanChildren()
	t.watcher.Remove(dir.Path)
	// TODO: remove recursive children...
}
func (t *Tree) openDir(dir *Node) error {
	if !dir.Info.IsDir() {
		return nil
	}
	err := dir.readChildren(t.sortingFunc)
	if err != nil {
		return err
	}
	t.watcher.Add(dir.Path)
	for _, ch := range dir.Children {
		t.searchIndex[ch.Info.Name()] = ch
	}
	return nil
}

func InitTree(dir string, sortingFunc NodeSortingFunc) (*Tree, <-chan NodeChange, error) {
	rootInfo, err := os.Lstat(dir)
	if err != nil {
		return nil, nil, err
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

	// Reading root directory children.
	err = root.readChildren(sortingFunc)
	if err != nil {
		return nil, nil, err
	}
	if len(root.Children) == 0 {
		return nil, nil, fmt.Errorf("Can't initialize on empty directory '%s'", dir)
	}

	// Initializing search index.
	searchIndex := map[string]*Node{}
	for _, ch := range root.Children {
		searchIndex[ch.Info.Name()] = ch
	}

	// Setup watcher for root directory
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
		searchIndex: searchIndex,
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
