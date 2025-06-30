package tree

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TreeTestSuite struct {
	suite.Suite
	tree *Tree
}

func (s *TreeTestSuite) SetupTest() {
	// Create tmp tmpdir files and dirs
	tmpdir := s.T().TempDir()
	rootDirName := "rootdir"
	rootPath := path.Join(tmpdir, rootDirName)
	createTestFileTree(
		s.T(),
		tmpdir,
		testnode{
			name:   rootDirName,
			isFile: false,
			children: []testnode{
				{
					name:     "empty_dir",
					isFile:   false,
					children: []testnode{},
				},
				{
					name:   "inner_dir",
					isFile: false,
					children: []testnode{
						{
							name:    "inner_file",
							isFile:  true,
							content: []byte("inner file content"),
						},
					},
				},
				{
					name:    "testfile.txt",
					isFile:  true,
					content: []byte("some content"),
				},
			},
		},
	)

	// Init Tree with this dir
	tree, _, err := InitTree(rootPath, defaultNodeSorting)
	s.Require().NoError(err)
	s.tree = tree
}

func (s *TreeTestSuite) TestInit() {
	s.Require().NotNil(s.tree.Root)
	s.Require().Len(s.tree.Root.Children, 3)
}
func (s *TreeTestSuite) TestMultiMarkSelected() {
	ok := s.tree.MarkSelectedChild()
	s.Require().True(ok)

	s.tree.SelectNextChild()

	ok = s.tree.MarkSelectedChild()
	s.Require().True(ok)

	s.Require().Len(s.tree.Marked, 2)
}
func (s *TreeTestSuite) TestMultiToggleMarkSelected() {
	// Marking
	ok := s.tree.ToggleMarkSelectedChild()
	s.Require().True(ok)
	s.Require().Len(s.tree.Marked, 1)

	s.tree.SelectNextChild()

	ok = s.tree.ToggleMarkSelectedChild()
	s.Require().True(ok)
	s.Require().Len(s.tree.Marked, 2)

	// Unmarking
	ok = s.tree.ToggleMarkSelectedChild()
	s.Require().True(ok)
	s.Require().Len(s.tree.Marked, 1)

	s.tree.SelectPreviousChild()

	ok = s.tree.ToggleMarkSelectedChild()
	s.Require().True(ok)
	s.Require().Len(s.tree.Marked, 0)
}
func (s *TreeTestSuite) TestMultiSelectCopy() {
	// Marking inner_dir and testfile.txt
	s.tree.SelectNextChild()
	ok := s.tree.MarkSelectedChild()
	s.Require().True(ok)
	s.tree.SelectNextChild()
	ok = s.tree.MarkSelectedChild()
	s.Require().True(ok)
	s.Require().Len(s.tree.Marked, 2)

	// Going top to empty_dir
	s.tree.CurrentDir.SelectFirst()
	err := s.tree.SetSelectedChildAsCurrent()
	s.Require().NoError(err)
	s.Require().Equal("empty_dir", s.tree.CurrentDir.Info.Name())

	// Copying marked
	err = s.tree.CopyMarkedToCurrentDir()
	s.Require().NoError(err)

	// note: not the cleanest way...
	s.tree.CurrentDir.readChildren(defaultNodeSorting)

	s.Require().Len(s.tree.CurrentDir.Children, 2)
	s.Require().Len(s.tree.Marked, 0)
}

func TestTreeTestSuite(t *testing.T) {
	suite.Run(t, new(TreeTestSuite))
}

type testnode struct {
	isFile   bool
	name     string
	content  []byte
	children []testnode
}

func createTestFileTree(tb testing.TB, dir string, node testnode) {
	tb.Helper()
	if node.isFile {
		f, err := os.Create(path.Join(dir, node.name))
		require.NoError(tb, err)
		defer f.Close()

		_, err = f.Write(node.content)
		require.NoError(tb, err)

	} else {
		newDir := path.Join(dir, node.name)
		err := os.Mkdir(newDir, os.ModePerm)
		require.NoError(tb, err)

		for _, child := range node.children {
			createTestFileTree(tb, newDir, child)
		}
	}
}
