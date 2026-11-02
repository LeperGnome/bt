package ui

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

type styledText struct {
	string
	lipgloss.Style
}

type multistyleText struct {
	textChunks []styledText
}

func (m *multistyleText) Len() int {
	// should upd on add?
	acc := 0
	for _, t := range m.textChunks {
		acc += utf8.RuneCountInString(t.string)
	}
	return acc
}

func (m *multistyleText) Copy() multistyleText {
	return multistyleText{
		textChunks: slices.Clone(m.textChunks),
	}
}

func (m *multistyleText) Add(text string, style lipgloss.Style) {
	m.textChunks = append(m.textChunks, styledText{text, style})
}

func (m *multistyleText) Append(mt multistyleText) {
	m.textChunks = append(m.textChunks, mt.textChunks...)
}

func (m *multistyleText) TrimTo(n int) {
	left := n
	lastText := 0
	for i := range m.textChunks {
		t := &m.textChunks[i]
		lastText = i
		newLeft := left - utf8.RuneCountInString(t.string)
		if newLeft <= 0 {
			t.string = string([]rune(t.string)[:max(0, left)])
			break
		}
		left = newLeft
	}

	m.textChunks = m.textChunks[:lastText+1]
}

func (m *multistyleText) Render() string {
	ss := make([]string, len(m.textChunks))
	for i, t := range m.textChunks {
		ss[i] = t.Style.Render(t.string)
	}

	return strings.Join(ss, "")
}
