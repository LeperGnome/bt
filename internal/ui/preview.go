package ui

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
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
	case "jpg", "jpeg":
		return halfBlockPreview // TODO: have to use it only in kitty / ghostty ...
	default:
		return plainTextPreview
	}

}

func tgpPreviewPNG(node *t.Node, height, width int, style Stylesheet) string {
	preview, err := tgpDirectChunks(node.Path)
	if err != nil {
		return err.Error()
	}
	return preview
}

func tgpDirectChunks(path string) (string, error) {
	res := []string{}
	data, err := os.ReadFile(path) // TODO
	if err != nil {
		return "", err
	}
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

func halfBlockPreview(node *t.Node, height, width int, style Stylesheet) string {
	preview := imageHalfBlockRepr(node.Path, height, width)
	return preview
}

func imageHalfBlockRepr(path string, height, width int) string {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	height -= 1 // For bottom resolution output
	height *= 2 // Due to most fonts w/h ratio.

	res := []string{}

	pxWidth := img.Bounds().Max.X
	pxHeight := img.Bounds().Max.Y

	sectorWidthRatio := int(math.Ceil(float64(pxWidth) / float64(width)))
	sectorHeightRatio := int(math.Ceil(float64(pxHeight) / float64(height)))

	sectorSide := max(sectorWidthRatio, sectorHeightRatio)

	width = pxWidth / sectorSide
	height = pxHeight / sectorSide

	if height&1 == 1 {
		height -= 1
	}

	pxCnt := uint32(sectorSide * sectorSide)

	// loop through sectors
	for sy := 0; sy < height; sy += 2 {
		line := []string{}
		for sx := range width {

			// loop through pixels within sector
			sums := [3]uint32{0, 0, 0}
			for y := sy * sectorSide; y < (sy+1)*sectorSide; y++ {
				for x := sx * (sectorSide); x < (sx+1)*sectorSide; x++ {
					c := img.At(x, y)
					r, g, b, _ := c.RGBA()
					sums[0] += r
					sums[1] += g
					sums[2] += b
				}
			}
			meansTop := [3]uint32{sums[0] / pxCnt, sums[1] / pxCnt, sums[2] / pxCnt}

			sums2 := [3]uint32{0, 0, 0}
			for y := (sy + 1) * sectorSide; y < (sy+2)*sectorSide; y++ {
				for x := sx * (sectorSide); x < (sx+1)*sectorSide; x++ {
					c := img.At(x, y)
					r, g, b, _ := c.RGBA()
					sums2[0] += r
					sums2[1] += g
					sums2[2] += b
				}
			}

			meansBottom := [3]uint32{sums2[0] / pxCnt, sums2[1] / pxCnt, sums2[2] / pxCnt}
			line = append(line, fmt.Sprintf(
				"\x1b[48;2;%d;%d;%dm\x1b[38;2;%d;%d;%dm▄\x1b[0m",
				meansTop[0]>>8,
				meansTop[1]>>8,
				meansTop[2]>>8,

				meansBottom[0]>>8,
				meansBottom[1]>>8,
				meansBottom[2]>>8,
			))
		}
		res = append(res, strings.Join(line, ""))
	}
	res = append(res, fmt.Sprintf("%dx%d", pxWidth, pxHeight))
	return strings.Join(res, "\n")
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
