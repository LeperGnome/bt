package ui

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	t "github.com/LeperGnome/bt/internal/tree"
)

type PreviewFunc = func(node *t.Node, height, width int, style Stylesheet) string

const kittyClearMedia = "\033_Ga=d\033\\"

func GetPreview(node *t.Node, height, width int, style Stylesheet) string {
	parts := strings.Split(node.Info.Name(), ".")
	ext := parts[len(parts)-1]
	f := getPreviewFunc(ext)

	preview := f(node, height, width, style)
	return kittyClearMedia + preview // TODO: my abstractions suck
}

func getPreviewFunc(fileType string) PreviewFunc {
	switch fileType {
	case "png":
		return tgpPreviewPNG // TODO: have to use it only in kitty / ghostty ...
	default:
		return plainTextPreview
	}

}

func tgpPreviewPNG(node *t.Node, height, width int, style Stylesheet) string {
	path64 := base64.StdEncoding.EncodeToString([]byte(node.Path))
	return fmt.Sprintf("\033_Ga=T,t=f,C=1,f=100;%s\033\\", path64)
}

func plainTextPreview(node *t.Node, height, width int, style Stylesheet) string {
	buf := make([]byte, previewTextBytesLimit) // TODO: fixed size buffer?
	n, err := node.ReadContent(buf, previewTextBytesLimit)
	if err != nil {
		return ""
	}
	content := buf[:n]

	contentStyle := style.PlainTextPreview.MaxWidth(width - 1) // -1 for border...

	var contentLines []string
	if !utf8.Valid(content) {
		contentLines = []string{binaryContentPlaceholder}
	} else {
		contentLines = strings.Split(string(content), "\n")
		contentLines = contentLines[:max(min(height, len(contentLines)), 0)]
	}

	return contentStyle.Render(strings.Join(contentLines, "\n"))
}
