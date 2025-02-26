package main

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"

	"github.com/LeperGnome/bt/internal/ui"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BtTestSuite struct {
	suite.Suite
	m  model
	tm *teatest.TestModel
}

func (s *BtTestSuite) SetupTest() {
	dir := s.T().TempDir()

	f, err := os.Create(path.Join(dir, "testfile.txt"))
	s.Require().NoError(err)
	defer f.Close()

	_, err = f.WriteString("some text")
	s.Require().NoError(err)

	m, err := newModel(dir, 5, ui.DefaultStylesheet, true)
	s.Require().NoError(err)

	tm := teatest.NewTestModel(s.T(), m, teatest.WithInitialTermSize(100, 100))

	s.T().Cleanup(func() {
		err := tm.Quit()
		s.Require().NoError(err)
	})

	s.m = m
	s.tm = tm
}

// TODO:
// Want to test:
// - navigation
// - file manipulations
//   - moving
//   - deleting
//   - renaming
//   - copying
//   - naming conflicts
//   - gg while moving/copying

func (s *BtTestSuite) TestExample() {
	teatest.WaitFor(s.T(), s.tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("testfile.txt"))
	})
}

func TestBtTestSuite(t *testing.T) {
	suite.Run(t, new(BtTestSuite))
}

func readBts(tb testing.TB, r io.Reader) []byte {
	tb.Helper()
	b, err := io.ReadAll(r)
	require.NoError(tb, err)
	return b
}
