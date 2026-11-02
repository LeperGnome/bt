package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/require"
)

func TestMTTrimLast(t *testing.T) {
	mt := multistyleText{}
	mt.Add("aaaaa", lipgloss.NewStyle())
	mt.Add("bbbbb", lipgloss.NewStyle())

	mt.TrimTo(7)

	require.Equal(t, 7, mt.Len())
	require.Equal(t, "aaaaabb", mt.Render())
}

func TestMTTrimMid(t *testing.T) {
	mt := multistyleText{}
	mt.Add("aaaaa", lipgloss.NewStyle())
	mt.Add("bbbbb", lipgloss.NewStyle())
	mt.Add("ccccc", lipgloss.NewStyle())

	mt.TrimTo(7)

	require.Equal(t, 7, mt.Len())
	require.Equal(t, "aaaaabb", mt.Render())
}
