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
	previewTextBytesLimit  int64 = 10_000
	previewMediaBytesLimit int64 = 1_000_000

	minHeight = 10
	minWidth  = 10

	arrow               = " <-"
	indentParent        = "│  "
	indentCurrent       = "├─ "
	indentCurrentLast   = "└─ "
	indentEmpty         = "   "
	emptydirContentName = "..."

	tooSmall                 = "too small =("
	binaryContentPlaceholder = "<binary content>"
	helpPreview              = "Press ? to toggle help"
)

type Renderer struct {
	Style       Stylesheet
	EdgePadding int
	offsetMem   int
	previewBuff [previewTextBytesLimit]byte // TODO: not needed?
}

func (r *Renderer) Render(s *state.State, winHeight, winWidth int) string {
	if winWidth < minWidth || winHeight < minHeight {
		return tooSmall
	}

	renderedHeading, headLen := r.renderHeading(s, winWidth)

	// section is half a screen, devided vertically
	// left for tree, right for file preview
	sectionWidth := int(math.Floor(0.5 * float64(winWidth)))

	renderedTree := r.renderTree(s.Tree, winHeight-headLen, sectionWidth)

	var rightPane string

	if s.HelpToggle {
		renderedHelp, helpLen := r.renderHelp(sectionWidth)
		renderedContent := r.renderSelectedFileContent(s.Tree, winHeight-headLen-helpLen, sectionWidth)
		rightPane = lipgloss.JoinVertical(lipgloss.Left, renderedHelp, renderedContent)
	} else {
		renderedContent := r.renderSelectedFileContent(s.Tree, winHeight-headLen, sectionWidth)
		rightPane = renderedContent
	}

	renderedTreeWithContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderedTree,
		rightPane,
	)

	return renderedHeading + "\n" + renderedTreeWithContent
}

func (r *Renderer) renderHeading(s *state.State, width int) (string, int) {
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

	rawPath := "> " + path

	finfo := fmt.Sprintf(
		"%s %s %v %s %s",
		r.Style.FinfoPermissions.Render(perm),
		r.Style.FinfoSep.Render("│"),
		r.Style.FinfoLastUpdated.Render(changeTime),
		r.Style.FinfoSep.Render("│"),
		r.Style.FinfoSize.Render(size),
	)

	header := []string{
		r.Style.SelectedPath.Render(rawPath) +
			strings.Repeat(
				" ",
				max(width-utf8.RuneCountInString(rawPath)-utf8.RuneCountInString(helpPreview), 0),
			) +
			r.Style.HelpMsg.Render(helpPreview),
		finfo,
		r.Style.OperationBar.Render(operationBar),
		r.Style.ErrBar.Render(s.ErrBuf),
	}
	return strings.Join(header, "\n"), len(header)
}

func (r *Renderer) renderHelp(width int) (string, int) {
	help := []string{
		"j / arr down   Select next child",
		"k / arr up     Select previous child",
		"h / arr left   Move up a dir",
		"l / arr right  Enter selected directory",
		"if / id	    Create file (if) / directory (id) in current directory",
		"d              Move selected child (then 'p' to paste)",
		"y              Copy selected child (then 'p' to paste)",
		"D              Delete selected child",
		"r              Rename selected child",
		"e              Edit selected file in $EDITOR",
		"gg             Go to top most child in current directory",
		"G              Go to last child in current directory",
		"H              Toggle hidden files in current directory",
		"enter          Collapse / expand selected directory",
		"esc            Clear error message / stop current operation",
		"q / ctrl+c     Exit",
	}
	return r.Style.
		HelpContent.
		MaxWidth(width).
		MarginRight(width).
		Render(strings.Join(help, "\n")), len(help) + 1 // +1 for border
}

func (r *Renderer) renderTree(tree *t.Tree, height, width int) string {
	renderedTreeLines, selectedRow := r.renderTreeFull(tree, width)
	croppedTreeLines := r.cropTree(renderedTreeLines, selectedRow, height)

	treeStyle := lipgloss.
		NewStyle().
		MaxWidth(width).
		MarginRight(width)

	return treeStyle.Render(strings.Join(croppedTreeLines, "\n"))
}

func (r *Renderer) renderSelectedFileContent(tree *t.Tree, height, width int) string {
	ch := tree.GetSelectedChild()
	if ch == nil {
		return "<no child>" // TODO
	}
	return GetPreview(ch, height, width, r.Style)
}

// Crops tree lines, such that current line is visible and view is consistent.
func (r *Renderer) cropTree(lines []string, currentLine int, height int) []string {
	linesLen := len(lines)

	// determining offset and limit based on selected row
	offset := r.offsetMem
	limit := linesLen

	// cursor is out for 'top' boundary
	if currentLine+1 > height+offset-r.EdgePadding {
		offset = max(min(currentLine+1-height+r.EdgePadding, linesLen-height), 0)
	}
	// cursor is out for 'bottom' boundary
	if currentLine < r.EdgePadding+offset {
		offset = max(currentLine-r.EdgePadding, 0)
	}
	r.offsetMem = offset
	limit = min(height+offset, linesLen)
	return lines[offset:limit]
}

// Returns lines as slice and index of selected line.
func (r *Renderer) renderTreeFull(tree *t.Tree, width int) ([]string, int) {
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

		if nameRuneCountNoStyle+indentRuneCount > width-6 { // 6 = len([]rune{"... <-"})
			name = string([]rune(name)[:max(0, width-indentRuneCount-6)]) + "..."
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

		if !node.ShowsHidden() && node.Info.IsDir() {
			repr += "◦"
		}

		if tree.GetSelectedChild() == node {
			repr += r.Style.TreeSelectionArrow.Render(arrow)
			currentLine = linen
		}
		lines = append(lines, repr)

		if node.Children != nil {
			// current directory is empty
			if len(node.Children) == 0 && tree.CurrentDir == node {
				emptyIndent := r.Style.TreeIndent.Render(parentIndent + indentCurrentLast)
				lines = append(lines, emptyIndent+emptydirContentName+r.Style.TreeSelectionArrow.Render(arrow))
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

var sizes = [...]string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}

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
