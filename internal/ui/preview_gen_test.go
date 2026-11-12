package ui

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	imgDir  = "./../../test/data/img/"
	reprDir = "./../../test/data/img/repr/"
	height  = 12
	width   = 12
)

type hbTest struct {
	imgPath, reprPath string
}

var hbTests = []hbTest{
	{"coltrane.jpg", "coltrane.repr"},
	{"lenna.png", "lenna.repr"},
}

func TestHalfBlockRepr(t *testing.T) {
	for _, tc := range hbTests {
		i, err := os.Open(imgDir + tc.imgPath)
		if err != nil {
			t.Error(err)
		}
		defer i.Close()

		o, err := os.ReadFile(reprDir + tc.reprPath)
		if err != nil {
			t.Error(err)
		}

		repr := imageHalfBlockRepr(i, height, width)
		if repr != string(o) {
			fmt.Printf("got:\n%s\nwant:\n%s\n", string(o), repr)
			t.Fatalf("Repr does not match for %s", tc.imgPath)
		}
	}
}

type lwTest struct {
	content       string
	width         int
	expectedSplit []string
}

var lwTests = []lwTest{
	{
		"0123456789",
		5,
		[]string{
			"01234",
			"56789",
		},
	},
	{
		"0123456789",
		3,
		[]string{
			"012",
			"345",
			"678",
			"9",
		},
	},
	{
		"0123456789",
		100,
		[]string{
			"0123456789",
		},
	},
	{
		"01234\n56789\n",
		3,
		[]string{
			"012",
			"34",
			"567",
			"89",
		},
	},
	{
		"	g", // it's \tg
		2,
		[]string{
			"  ",
			"  ",
			"g",
		},
	},
}

func TestLineWrap(t *testing.T) {
	for _, tc := range lwTests {
		require.Equal(
			t,
			tc.expectedSplit,
			toWrappedLines(tc.content, tc.width),
		)
	}
}
