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
		return tgpPreview // TODO: have to use it only in kitty / ghostty ...
	default:
		return plainTextPreview
	}

}

func tgpPreview(node *t.Node, height, width int, style Stylesheet) string {
	buf := make([]byte, previewTextBytesLimit) // TODO: fixed size buffer?
	n, err := node.ReadContent(buf, previewMediaBytesLimit)
	if err != nil {
		return ""
	}
	chunkSize := 1024

	data64 := make([]byte, base64.StdEncoding.EncodedLen(n))
	base64.StdEncoding.Encode(data64, buf[:n])

	media := []string{}
	chunkN := 0

	// TODO not working
	for {
		if len(data64) > chunkSize*(chunkN+1) && chunkN == 0 {
			media = append(media, fmt.Sprintf("\033_GC=1,f=100,m=1;%s\033\\", data64[chunkSize*chunkN:chunkSize*(chunkN+1)]))
			chunkN++
		} else if len(data64) > chunkSize*(chunkN+1) {
			media = append(media, fmt.Sprintf("\033_Gm=1;%s\033\\", data64[chunkSize*chunkN:chunkSize*(chunkN+1)]))
			chunkN++
		} else {
			media = append(media, fmt.Sprintf("\033_Gm=0;%s\033\\", data64[chunkSize*chunkN:]))
			break
		}
	}
	return strings.Join(media, "\n")
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
