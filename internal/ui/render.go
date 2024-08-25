package ui

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"

	t "github.com/LeperGnome/bt/internal/tree"
)

const (
	previewBytesLimit int64 = 10_000
	minHeight               = 10
	minWidth                = 10
)

type Renderer struct {
	EdgePadding int
	offsetMem   int
}

func (r *Renderer) Render(tree *t.Tree, winHeight, winWidth int) string {
	if winWidth < minWidth || winHeight < minHeight {
		return "too small =(\n"
	}

	// section is half a screen, devided vertically
	// left for tree, right for file preview
	sectionWidth := int(math.Floor(0.5 * float64(winWidth)))

	// rendering tree
	renderedTree, selectedRow := r.renderTree(tree, sectionWidth)
	croppedTree := r.cropTree(renderedTree, selectedRow, winHeight)

	treeStyle := lipgloss.
		NewStyle().
		MaxWidth(sectionWidth)
	renderedStyledTree := treeStyle.Render(strings.Join(croppedTree, "\n"))

	// rendering file content
	content := make([]byte, previewBytesLimit)
	n, err := tree.ReadSelectedChildContent(content, previewBytesLimit)
	if err != nil {
		return renderedStyledTree
	}
	content = content[:n]

	leftMargin := sectionWidth - lipgloss.Width(renderedStyledTree)
	contentStyle := lipgloss.
		NewStyle().
		Italic(true).
		MarginLeft(leftMargin).
		MaxWidth(sectionWidth + leftMargin - 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true)

	var contentLines []string
	if !utf8.Valid(content) {
		contentLines = []string{"<binary content>"}
	} else {
		contentLines = strings.Split(string(content), "\n")
		contentLines = contentLines[:max(min(winHeight, len(contentLines)), 0)]
	}
	renderedStyledTree = lipgloss.JoinHorizontal(
		0,
		renderedStyledTree,
		contentStyle.Render(strings.Join(contentLines, "\n")),
	)
	return renderedStyledTree
}

// Crops tree lines, such that current line is visible and view is consistent.
func (r *Renderer) cropTree(lines []string, currentLine int, windowHeight int) []string {
	linesLen := len(lines)

	// determining offset and limit based on selected row
	offset := r.offsetMem
	limit := linesLen
	if windowHeight > 0 {
		// cursor is out for 'top' boundary
		if currentLine+1 > windowHeight+offset-r.EdgePadding {
			offset = min(currentLine+1-windowHeight+r.EdgePadding, linesLen-windowHeight)
		}
		// cursor is out for 'bottom' boundary
		if currentLine < r.EdgePadding+offset {
			offset = max(currentLine-r.EdgePadding, 0)
		}
		r.offsetMem = offset
		limit = min(windowHeight+offset, linesLen)
	}
	return lines[offset:limit]
}

// Returns lines as slice and index of selected line.
func (r *Renderer) renderTree(tree *t.Tree, widthLim int) ([]string, int) {
	linen := -1
	currentLine := 0

	type stackEl struct {
		*t.Node
		int
	}
	lines := []string{}
	s := newStack(stackEl{tree.Root, 0})

	for s.Len() > 0 {
		el := s.Pop()
		linen += 1

		node := el.Node
		depth := el.int

		if node == nil {
			continue
		}

		name := node.Info.Name()
		nameRuneCountNoStyle := utf8.RuneCountInString(name)
		indent := strings.Repeat("  ", depth)
		indentRuneCount := utf8.RuneCountInString(indent)

		// TODO: probably bug here
		if nameRuneCountNoStyle+indentRuneCount > widthLim-6 { // 6 = len([]rune{"... <-"})
			name = string([]rune(name)[:max(0, widthLim-indentRuneCount-6)]) + "..."
		}

		if node.Info.IsDir() {
			name = color.BlueString(node.Info.Name())
		}
		if tree.Marked == node {
			s := lipgloss.
				NewStyle().
				Background(lipgloss.Color("#3C3C3C"))
			name = s.Render(name)
		}

		repr := indent + name

		if tree.GetSelectedChild() == node {
			repr += color.YellowString(" <-")
			currentLine = linen
		}
		lines = append(lines, repr)

		if node.Children != nil {
			// current directory is empty
			if len(node.Children) == 0 && tree.CurrentDir == node {
				lines = append(lines, strings.Repeat("  ", depth+1)+"..."+color.YellowString(" <-"))
				currentLine = linen + 1
			}
			for i := len(node.Children) - 1; i >= 0; i-- {
				ch := node.Children[i]
				s.Push(stackEl{ch, depth + 1})
			}
		}
	}
	return lines, currentLine
}

type stack[T any] struct {
	items []T
}

func (s *stack[T]) Push(el ...T) {
	s.items = append(s.items, el...)
}
func (s *stack[_]) Len() int {
	return len(s.items)
}
func (s *stack[T]) Pop() T {
	el := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return el
}

func newStack[T any](els ...T) stack[T] {
	s := stack[T]{}
	s.Push(els...)
	return s
}
