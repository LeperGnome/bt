package ui

import (
	"fmt"
	"os"
	"testing"
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
