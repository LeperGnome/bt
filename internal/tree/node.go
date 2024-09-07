package tree

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type NodeSortingFunc func(a, b *Node) int

type Node struct {
	Path             string
	Info             fs.FileInfo
	Children         []*Node // nil - not read or it's a file
	Parent           *Node
	selectedChildIdx int
	contentCache     []byte
}

func (n *Node) SelectLast() {
	// can I just check for len here?
	if n.Children != nil && len(n.Children) > 0 {
		n.selectedChildIdx = len(n.Children) - 1
	}
}
func (n *Node) SelectFirst() {
	n.selectedChildIdx = 0
}
func (n *Node) readChildren(sortFunc NodeSortingFunc) error {
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
				Path:     filepath.Join(n.Path, chInfo.Name()),
				Info:     chInfo,
				Children: nil,
				Parent:   n,
			}
		}
		chNodes = append(chNodes, childToAdd)
	}
	slices.SortFunc(chNodes, sortFunc)
	n.Children = chNodes

	// updateing selected child index if it's out of bounds after update
	n.selectedChildIdx = max(min(n.selectedChildIdx, len(n.Children)-1), 0)
	return nil
}
func (n *Node) orphanChildren() {
	n.Children = nil
}

func defaultNodeSorting(a, b *Node) int {
	// dirs first
	if a.Info.IsDir() != b.Info.IsDir() {
		if a.Info.IsDir() {
			return -1
		} else {
			return 1
		}
	}
	return strings.Compare(strings.ToLower(a.Info.Name()), strings.ToLower(b.Info.Name()))
}
