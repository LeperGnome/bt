package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

const NotSelected = -1

type Node struct {
	Path     string
	Info     fs.FileInfo
	Children []Node // nil - not read or it's a file
	Parent   *Node
	Selected int
	smem     int
}

func (n *Node) ReadChildren() error {
	if !n.Info.IsDir() {
		return nil
	}
	children, err := os.ReadDir(n.Path)
	if err != nil {
		return err
	}
	chNodes := make([]Node, 0, len(children))
	for _, ch := range children {
		chInfo, err := ch.Info()
		if err != nil {
			return err
		}
		chNodes = append(chNodes, Node{
			Path:     filepath.Join(n.Path, chInfo.Name()),
			Info:     chInfo,
			Children: nil,
			Parent:   n,
			Selected: NotSelected,
		})
	}
	n.Children = chNodes
	return nil
}

func (n *Node) OrphanChildren() {
	n.Children = nil
	n.Selected = NotSelected
}

func InitRoot(dir string) (*Node, error) {
	rootInfo, err := os.Lstat(dir)
	if err != nil {
		return nil, err
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}
	root := &Node{
		Path:     dir,
		Info:     rootInfo,
		Parent:   nil,
		Children: []Node{},
	}

	err = root.ReadChildren()
	if err != nil {
		return nil, err
	}
	return root, nil
}

func (n *Node) View() string {
	b := strings.Builder{}
	printNode(&b, n, 0, false)
	return b.String()
}

func PrintNode(node *Node) {
	printNode(os.Stdout, node, 0, false)
}

func printNode(
	b io.Writer,
	node *Node,
	depth int,
	marked bool,
) {
	if node == nil {
		return
	}
	name := node.Info.Name()
	if node.Info.IsDir() {
		name = color.BlueString(node.Info.Name())
	}
	repr := strings.Repeat("  ", depth) + name
	if marked {
		repr += color.YellowString(" <-")
	}
	repr += "\n"
	_, err := b.Write([]byte(repr))
	if err != nil {
		panic("could not write node")
	}

	if node.Children != nil {
		for cidx, ch := range node.Children {
			printNode(b, &ch, depth+1, cidx == node.Selected)
		}
	}
}
