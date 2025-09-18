package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	flag "github.com/spf13/pflag"

	"github.com/LeperGnome/bt/internal/config"
	"github.com/LeperGnome/bt/internal/state"
	"github.com/LeperGnome/bt/internal/tree"
	ui "github.com/LeperGnome/bt/internal/ui"
)

type model struct {
	window ui.Dimentions

	appState *state.State
	renderer *ui.Renderer
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		listenFSEvents(m.appState.NodeChanges),
		listenPreviewReady(m.renderer.PreviewDoneChan),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.window = ui.Dimentions{Height: msg.Height, Width: msg.Width}
	case tea.KeyMsg:
		return m, m.appState.ProcessKey(msg)
	case tree.NodeChange:
		m.renderer.RemovePreviewCache(msg.Path)
		m.appState.ProcessNodeChange(msg)
		return m, listenFSEvents(m.appState.NodeChanges)
	case ui.Preview:
		m.renderer.SetPreviewCache(msg)
		return m, listenPreviewReady(m.renderer.PreviewDoneChan)
	}
	return m, nil
}
func (m model) View() string {
	return m.renderer.Render(m.appState, m.window)
}

func newModel(
	root string,
	style ui.Stylesheet,
	padding int,
	filePreview bool,
	highlightCurrentIndent bool,
) (model, error) {
	s, err := state.InitState(root)
	if err != nil {
		return model{}, err
	}
	renderer := ui.NewRenderer(style, padding, filePreview, highlightCurrentIndent)
	return model{
		appState: s,
		renderer: renderer,
	}, nil
}

func listenFSEvents(eventsChan <-chan tree.NodeChange) tea.Cmd {
	return func() tea.Msg {
		return <-eventsChan
	}
}

func listenPreviewReady(previewChan <-chan ui.Preview) tea.Cmd {
	return func() tea.Msg {
		return <-previewChan
	}
}

func main() {
	flag.UintP("padding", "p", 5, "Edge padding for top and bottom")
	flag.BoolP("in_place_render", "i", false, "In-place render (without alternate screen)")
	flag.Bool("file_preview", true, "Enable file previews")
	flag.Bool("highlight_indent", true, "Highlight current indent")

	flag.Parse()

	conf := config.GetConfig(flag.CommandLine)

	rootPath := flag.Arg(0)
	if rootPath == "" {
		rootPath = "."
	}

	m, err := newModel(
		rootPath,
		ui.DefaultStylesheet,
		conf.Padding,
		conf.FilePreview,
		conf.HighlightIndent,
	)
	if err != nil {
		fmt.Printf("Error on init: %v", err)
		os.Exit(1)
	}

	opts := []tea.ProgramOption{}
	if !conf.InPlaceRender {
		opts = append(opts, tea.WithAltScreen())
	}

	p := tea.NewProgram(m, opts...)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
