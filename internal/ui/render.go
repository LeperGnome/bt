package ui

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"

	"github.com/LeperGnome/bt/internal/state"
	t "github.com/LeperGnome/bt/internal/tree"
	"github.com/LeperGnome/bt/pkg/stack"
)

const (
	previewBytesLimit   int64 = 10_000
	minHeight                 = 10
	minWidth                  = 10
	indentParent              = "│  "
	indentCurrent             = "├─ "
	indentCurrentLast         = "└─ "
	indentEmpty               = "   "
	emptydirContentName       = "..."
)

type Renderer struct {
	Style       Stylesheet
	EdgePadding int
	offsetMem   int
	contentBuf  [previewBytesLimit]byte
}

func (r *Renderer) Render(s *state.State, winHeight, winWidth int) string {
	renderedHeading, headLen := r.renderHeading(s)
	renderedTree := r.renderTreeWithContent(s.Tree, winHeight-headLen, winWidth)

	return renderedHeading + "\n" + renderedTree
}

func (r *Renderer) renderHeading(s *state.State) (string, int) {
	selected := s.Tree.GetSelectedChild()

	// NOTE: special case for empty dir
	path := s.Tree.CurrentDir.Path + "/..."
	changeTime := "--"
	size := "0 B"
	perm := "--"

	if selected != nil {
		path = selected.Path
		changeTime = selected.Info.ModTime().Format(time.RFC822)
		size = formatSize(float64(selected.Info.Size()), 1024.0)
		perm = selected.Info.Mode().String()
	}

	markedPath := ""
	if s.Tree.Marked != nil {
		markedPath = s.Tree.Marked.Path
	}

	operationBar := fmt.Sprintf(": %s", s.OpBuf.Repr())
	if markedPath != "" {
		operationBar += fmt.Sprintf(" [%s]", markedPath)
	}

	if s.OpBuf.IsInput() {
		operationBar += fmt.Sprintf(" │ %s │", r.Style.OperationBarInput.Render(string(s.InputBuf)))
	}

	header := []string{
		r.Style.SelectedPath.Render("> " + path),
		fmt.Sprintf(
			"%s %s %v %s %s",
			r.Style.FinfoPermissions.Render(perm),
			r.Style.FinfoSep.Render("│"),
			r.Style.FinfoLastUpdated.Render(changeTime),
			r.Style.FinfoSep.Render("│"),
			r.Style.FinfoSize.Render(size),
		),
		r.Style.OperationBar.Render(operationBar),
		r.Style.ErrBar.Render(s.ErrBuf),
	}
	return strings.Join(header, "\n"), len(header)
}

func (r *Renderer) renderTreeWithContent(tree *t.Tree, winHeight, winWidth int) string {
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
	n, err := tree.ReadSelectedChildContent(r.contentBuf[:], previewBytesLimit)
	if err != nil {
		return renderedStyledTree
	}
	content := r.contentBuf[:n]

	leftMargin := sectionWidth - lipgloss.Width(renderedStyledTree)
	contentStyle := r.Style.ContentPreview.
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		MarginLeft(leftMargin).
		MaxWidth(sectionWidth + leftMargin - 1)

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
			offset = max(min(currentLine+1-windowHeight+r.EdgePadding, linesLen-windowHeight), 0)
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
		string
		bool
	}
	lines := []string{}
	s := stack.NewStack(stackEl{tree.Root, "", false})

	for s.Len() > 0 {
		el := s.Pop()
		linen += 1

		node := el.Node
		isLast := el.bool
		parentIndent := el.string

		var indent string
		if node == tree.Root {
			indent = ""
		} else if isLast {
			indent = parentIndent + indentCurrentLast
			parentIndent = parentIndent + indentEmpty
		} else {
			indent = parentIndent + indentCurrent
			parentIndent = parentIndent + indentParent
		}

		if node == nil {
			continue
		}

		name := node.Info.Name()
		nameRuneCountNoStyle := utf8.RuneCountInString(name)
		indentRuneCount := utf8.RuneCountInString(indent)

		if nameRuneCountNoStyle+indentRuneCount > widthLim-6 { // 6 = len([]rune{"... <-"})
			name = string([]rune(name)[:max(0, widthLim-indentRuneCount-6)]) + "..."
		}

		indent = r.Style.TreeIndent.Render(indent)

		if node.Info.IsDir() {
			name = r.Style.TreeDirecotryName.Render(name)
		} else if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
			name = r.Style.TreeLinkName.Render(name)
		} else {
			name = r.Style.TreeRegularFileName.Render(name)
		}

		if tree.Marked == node {
			name = r.Style.TreeMarkedNode.Render(name)
		}

		repr := indent + name

		if tree.GetSelectedChild() == node {
			repr += r.Style.TreeSelectionArrow.Render(" <-")
			currentLine = linen
		}
		lines = append(lines, repr)

		if node.Children != nil {
			// current directory is empty
			if len(node.Children) == 0 && tree.CurrentDir == node {
				emptyIndent := r.Style.TreeIndent.Render(parentIndent + indentCurrentLast)
				lines = append(lines, emptyIndent+emptydirContentName+r.Style.TreeSelectionArrow.Render(" <-"))
				currentLine = linen + 1
			}
			for i := len(node.Children) - 1; i >= 0; i-- {
				ch := node.Children[i]
				s.Push(stackEl{ch, parentIndent, i == len(node.Children)-1})
			}
		}
	}
	return lines, currentLine
}

var sizes = [...]string{"b", "Kb", "Mb", "Gb", "Tb", "Pb", "Eb"}

func formatSize(s float64, base float64) string {
	unitsLimit := len(sizes)
	i := 0
	for s >= base && i < unitsLimit {
		s = s / base
		i++
	}
	f := "%.0f %s"
	if i > 1 {
		f = "%.2f %s"
	}
	return fmt.Sprintf(f, s, sizes[i])
}
