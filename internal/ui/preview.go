package ui

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	t "github.com/LeperGnome/bt/internal/tree"
)

type PreviewFunc = func(node *t.Node, height, width int, style Stylesheet) string

const (
	kittyClearMedia = "\033_Ga=d\033\\"
	tgpChunkSize    = 2048
)

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
	buf := make([]byte, previewMediaBytesLimit)
	n, err := node.ReadContent(buf, previewTextBytesLimit)
	if err != nil {
		return err.Error()
	}
	data := buf[:n]

	preview, err := tgpDirectChunks(data)
	if err != nil {
		return err.Error()
	}

	return preview
}

func tgpDirectChunks(data []byte) (string, error) {
	res := []string{}
	data64 := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(data64, data)

	chunk := 0

	for {
		if chunk == 0 && (chunk+1)*tgpChunkSize < len(data64) {
			chunkData := fmt.Sprintf(
				"\033_Ga=T,C=1,f=100,m=1;%s\033\\",
				string(data64[chunk*tgpChunkSize:(chunk+1)*tgpChunkSize]),
			)
			res = append(res, chunkData)
		} else if (chunk+1)*tgpChunkSize < len(data64) {
			chunkData := fmt.Sprintf(
				"\033_Gm=1;%s\033\\",
				string(data64[chunk*tgpChunkSize:(chunk+1)*tgpChunkSize]),
			)
			res = append(res, chunkData)
		} else if chunk == 0 {
			chunkData := fmt.Sprintf(
				"\033_Ga=T,C=1,f=100;%s\033\\",
				string(data64),
			)
			res = append(res, chunkData)
			break
		} else {
			chunkData := fmt.Sprintf(
				"\033_Gm=0;%s\033\\",
				string(data64[chunk*tgpChunkSize:]),
			)
			res = append(res, chunkData)
			break
		}
		chunk++
	}

	return strings.Join(res, ""), nil
}

func plainTextPreview(node *t.Node, height, width int, style Stylesheet) string {
	buf := make([]byte, previewTextBytesLimit)
	n, err := node.ReadContent(buf, previewTextBytesLimit)
	if err != nil {
		return err.Error()
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
