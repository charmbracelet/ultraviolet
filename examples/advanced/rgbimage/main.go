package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/iterm2"
)

func main() {
	t := uv.DefaultTerminal()

	if err := t.Start(); err != nil {
		log.Fatalln("error starting terminal:", err)
	}
	defer t.Stop()

	scr := t.Screen()
	scr.EnterAltScreen()
	// scr.ShowCursor()

	var keypress string
	var cellWidth, cellHeight int

	tmpDir, _ := os.MkdirTemp("", "uv-rgbimage-example")
	defer os.RemoveAll(tmpDir)

	createEncodedImage := func(w, h int, col color.Color) string {
		img := createImage(w*cellWidth, h*cellHeight, col)
		png := imagePng(img)
		if err := os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("img_%d_%d.png", w, h)), png, 0o644); err != nil {
			log.Println("error writing png file:", err)
		}
		return dataBase64(png)
	}

	var imgXOffset, imgYOffset int
	display := func() {
		screen.Clear(scr)
		text := fmt.Sprintf("Press any key to see its value. Press 'ctrl+c' to quit.\n\nLast key pressed: %s", keypress)
		nlines := strings.Count(text, "\n") + 1
		ss := uv.NewStyledString(text)
		bounds := scr.Bounds()
		if bounds.Empty() {
			return
		}

		center := uv.Rect(
			(bounds.Max.X-ss.UnicodeWidth())/2,
			(bounds.Max.Y-nlines)/2,
			ss.UnicodeWidth(),
			nlines,
		)
		ss.Draw(scr, center)
		scr.Render()

		const imgW, imgH = 10, 5
		col := ansi.IndexedColor(rand.Intn(256))
		encodedPngImg := createEncodedImage(imgW, imgH, col)
		centerImg := uv.Rect(
			0+imgXOffset,
			0+imgYOffset,
			// ((bounds.Max.X-imgW)/2)+imgXOffset,
			// ((bounds.Max.Y-imgH)/2)+imgYOffset,
			imgW,
			imgH,
		)

		log.Printf("center image rect: %+v", centerImg)

		for y := centerImg.Min.Y; y < centerImg.Max.Y; y++ {
			log.Printf("setting cell at (%d, %d)", centerImg.Min.X, y)
			var content string
			if y == centerImg.Min.Y {
				content = ansi.ITerm2(iterm2.File{
					Name:              "img.png",
					Width:             iterm2.Cells(imgW),
					Height:            iterm2.Cells(imgH),
					Content:           []byte(encodedPngImg),
					Inline:            true,
					IgnoreAspectRatio: true,
					DoNotMoveCursor:   false,
				}) +
					// ansi.CursorUp(imgH-1) +
					// ansi.CursorForward(imgW+1)
					ansi.CursorPosition(centerImg.Min.X+imgW+1, centerImg.Min.Y+1)
				log.Printf("content for cell at (%d, %d): %q", centerImg.Min.X, y, content)
			} else {
				content = ansi.CursorForward(imgW)
			}

			scr.SetCell(centerImg.Min.X, y, &uv.Cell{
				Content: content,
				Width:   imgW,
			})
		}

		scr.Render()
		scr.Flush()
	}

	display()

loop:
	for ev := range t.Events() {
		switch e := ev.(type) {
		case uv.PixelSizeEvent:
			cellWidth = e.Width / scr.Bounds().Dx()
			cellHeight = e.Height / scr.Bounds().Dy()
		case uv.WindowSizeEvent:
			scr.Resize(e.Width, e.Height)
		case uv.KeyPressEvent:
			switch e.String() {
			case "ctrl+c":
				break loop
			case "up":
				imgYOffset--
			case "down":
				imgYOffset++
			case "left":
				imgXOffset--
			case "right":
				imgXOffset++
			}
			keypress = e.String()
		}

		display()
	}

	screen.Clear(scr)
	scr.Render()
	scr.Reset()
	scr.Flush()
}

func createImage(w, h int, col color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, col)
		}
	}
	return img
}

func imageRgba(img image.Image) []byte {
	var buf bytes.Buffer
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			buf.WriteByte(byte(r >> 8))
			buf.WriteByte(byte(g >> 8))
			buf.WriteByte(byte(b >> 8))
			buf.WriteByte(byte(a >> 8))
		}
	}
	return buf.Bytes()
}

func imagePng(img image.Image) []byte {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func dataBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func init() {
	f, err := os.OpenFile("uv_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalln("error opening debug log file:", err)
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
