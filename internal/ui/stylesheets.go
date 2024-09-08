package ui

import "github.com/charmbracelet/lipgloss"

type Stylesheet struct {
	SelectedPath lipgloss.Style

	FinfoPermissions lipgloss.Style
	FinfoLastUpdated lipgloss.Style
	FinfoSize        lipgloss.Style
	FinfoSep         lipgloss.Style

	OperationBar      lipgloss.Style
	OperationBarInput lipgloss.Style

	ErrBar lipgloss.Style

	TreeRegularFileName lipgloss.Style
	TreeDirecotryName   lipgloss.Style
	TreeLinkName        lipgloss.Style
	TreeMarkedNode      lipgloss.Style
	TreeSelectionArrow  lipgloss.Style
	TreeIndent          lipgloss.Style
	TreeSearchResult    lipgloss.Style

	ContentPreview lipgloss.Style
}

var DefaultStylesheet = Stylesheet{
	SelectedPath: lipgloss.NewStyle().Foreground(lipgloss.Color("#74AC6D")),

	FinfoPermissions: lipgloss.NewStyle().Foreground(lipgloss.Color("#ACA46D")),
	FinfoLastUpdated: lipgloss.NewStyle().Foreground(lipgloss.Color("#E6E6E6")),
	FinfoSize:        lipgloss.NewStyle().Foreground(lipgloss.Color("#E6E6E6")),
	FinfoSep:         lipgloss.NewStyle().Foreground(lipgloss.Color("#2b2b2b")),

	OperationBar:      lipgloss.NewStyle().Foreground(lipgloss.Color("#E6E6E6")),
	OperationBarInput: lipgloss.NewStyle().Background(lipgloss.Color("#3C3C3C")),

	ErrBar: lipgloss.NewStyle().Foreground(lipgloss.Color("#AC6D74")),

	TreeRegularFileName: lipgloss.NewStyle().Foreground(lipgloss.Color("#E6E6E6")),
	TreeDirecotryName:   lipgloss.NewStyle().Foreground(lipgloss.Color("#6D74AC")),
	TreeLinkName:        lipgloss.NewStyle().Foreground(lipgloss.Color("#6DACA4")),
	TreeMarkedNode: lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.InnerHalfBlockBorder()).
		Background(lipgloss.Color("#2b2b2b")),
	TreeSelectionArrow: lipgloss.NewStyle().Foreground(lipgloss.Color("#ACA46D")),
	TreeIndent:         lipgloss.NewStyle().Foreground(lipgloss.Color("#2b2b2b")),
	TreeSearchResult:   lipgloss.NewStyle().Background(lipgloss.Color("#4b3a57")),

	ContentPreview: lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#B8BCD6")),
}
