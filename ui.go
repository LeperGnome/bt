package main

import (
	"io"
	"strings"

	"github.com/fatih/color"
)

func Render(tree *Tree) (string, error) {
	b := strings.Builder{}
	err := renderNode(&b, tree.Root, 0, false)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func renderNode(
	b io.Writer,
	node *Node,
	depth int,
	marked bool,
) error {
	if node == nil {
		return nil
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
		return err
	}

	if node.Children != nil {
		for cidx, ch := range node.Children {
			renderNode(b, &ch, depth+1, cidx == node.Selected)
		}
	}
	return nil
}
