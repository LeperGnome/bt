package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

const NotSelected = -1

type Tree struct {
	Root       *Node
	CurrentDir *Node
}

func (t *Tree) GetSelectedChild() *Node {
	return t.CurrentDir.Children[t.CurrentDir.SelectedChildIdx]
}
func (t *Tree) SelectNextChild() {
	if t.CurrentDir.SelectedChildIdx < len(t.CurrentDir.Children)-1 {
		t.CurrentDir.SelectedChildIdx += 1
		t.CurrentDir.smem += 1
	}
}
func (t *Tree) SelectPreviousChild() {
	if t.CurrentDir.SelectedChildIdx > 0 {
		t.CurrentDir.SelectedChildIdx -= 1
		t.CurrentDir.smem -= 1
	}
}
func (t *Tree) SetSelectedChildAsCurrent() error {
	selectedChild := t.GetSelectedChild()
	if selectedChild.Children == nil {
		err := selectedChild.ReadChildren()
		if err != nil {
			return err
		}
	}
	if len(selectedChild.Children) > 0 {
		t.CurrentDir = selectedChild
		t.CurrentDir.SelectedChildIdx = t.CurrentDir.smem
	}
	return nil
}
func (t *Tree) SetParentAsCurrent() {
	if t.CurrentDir.Parent != nil {
		currentName := t.CurrentDir.Info.Name()
		// setting parent current to match the directory, that we're leaving
		newParentIdx := slices.IndexFunc(t.CurrentDir.Parent.Children, func(n *Node) bool { return n.Info.Name() == currentName })
		t.CurrentDir.Parent.SelectedChildIdx = newParentIdx
		t.CurrentDir.Parent.smem = newParentIdx

		t.CurrentDir.SelectedChildIdx = NotSelected
		t.CurrentDir = t.CurrentDir.Parent
	}
}
func (t *Tree) CollapseOrExpandSelected() error {
	selectedChild := t.GetSelectedChild()
	if selectedChild.Children != nil {
		selectedChild.orphanChildren()
	} else {
		err := selectedChild.ReadChildren()
		if err != nil {
			return err
		}
	}
	return nil
}

func InitTree(dir string) (Tree, error) {
	var tree Tree
	rootInfo, err := os.Lstat(dir)
	if err != nil {
		return tree, err
	}
	if !rootInfo.IsDir() {
		return tree, fmt.Errorf("%s is not a directory", dir)
	}
	root := &Node{
		Path:     dir,
		Info:     rootInfo,
		Parent:   nil,
		Children: []*Node{},
	}

	err = root.ReadChildren()
	if err != nil {
		return tree, err
	}
	if len(root.Children) == 0 {
		return tree, fmt.Errorf("Can't initialize on empty directory '%s'", dir)
	}
	tree = Tree{Root: root, CurrentDir: root}
	return tree, nil
}

type Node struct {
	Path             string
	Info             fs.FileInfo
	Children         []*Node // nil - not read or it's a file
	Parent           *Node
	SelectedChildIdx int
	smem             int
}

func (n *Node) ReadChildren() error {
	if !n.Info.IsDir() {
		return nil
	}
	children, err := os.ReadDir(n.Path)
	if err != nil {
		return err
	}
	chNodes := []*Node{}

	for _, ch := range children {
		ch := ch
		chInfo, err := ch.Info()
		if err != nil {
			return err
		}
		// Looking if child already exist
		var childToAdd *Node
		if n.Children != nil {
			for _, ech := range n.Children {
				if ech.Info.Name() == chInfo.Name() {
					childToAdd = ech
					break
				}
			}
		}
		if childToAdd == nil {
			childToAdd = &Node{
				Path:             filepath.Join(n.Path, chInfo.Name()),
				Info:             chInfo,
				Children:         nil,
				Parent:           n,
				SelectedChildIdx: NotSelected,
			}
		}
		chNodes = append(chNodes, childToAdd)
	}
	n.Children = chNodes
	return nil
}
func (n *Node) orphanChildren() {
	n.Children = nil
	n.SelectedChildIdx = NotSelected
}
