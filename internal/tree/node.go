package tree

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type NodeSortingFunc func(a, b *Node) int

type Node struct {
	Path     string
	Info     fs.FileInfo
	Children []*Node // nil - not read or it's a file
	Parent   *Node

	selectedChildIdx int
	showHidden       bool
	showHiddenDefault       bool
}

func (n *Node) SelectLast() {
	// can I just check for len here?
	if len(n.Children) > 0 {
		n.selectedChildIdx = len(n.Children) - 1
	}
}
func (n *Node) ShowsHiddenNonDefault() bool {
	return n.showHidden != n.showHiddenDefault
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
		chInfo, err := ch.Info()
		if err != nil {
			return err
		}
		// Skipping hidden node
		if !n.showHidden && strings.HasPrefix(chInfo.Name(), ".") {
			continue
		}
		// Looking if child already exist, so i'm keeping it's read children intact
		var childToAdd *Node
		if n.Children != nil {
			for _, ech := range n.Children {
				if ech.Info.Name() == chInfo.Name() {
					childToAdd = ech
					childToAdd.Info = chInfo // updating info in case file was changed
					break
				}
			}
		}
		if childToAdd == nil {
			childToAdd = NewNode(
				filepath.Join(n.Path, chInfo.Name()),
				chInfo,
				n,
				n.showHiddenDefault,
			)
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
func (n *Node) ReadContent(buf []byte, limit int64) (int, error) {
	if !n.Info.Mode().IsRegular() {
		return 0, fmt.Errorf("file not selected or is irregular")
	}
	f, err := os.Open(n.Path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	limitedReader := io.LimitReader(f, limit)
	k, err := limitedReader.Read(buf)
	if err != nil {
		return 0, err
	}
	return k, nil
}

func NewNode(path string, info fs.FileInfo, parent *Node, showHidden bool) *Node {
	return &Node{
		Path:       path,
		Info:       info,
		Children:   nil,
		Parent:     parent,
		showHidden: showHidden,
		showHiddenDefault: showHidden,
	}
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
