package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	_ "image/jpeg" // Register JPEG format

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/iterm2"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/charmbracelet/x/ansi/sixel"
	"github.com/charmbracelet/x/mosaic"
)

type imageEncoding uint8

const (
	blocksEncoding imageEncoding = iota + 1
	sixelEncoding
	itermEncoding
	kittyEncoding

	unknownEncoding = 0
)

func (e imageEncoding) String() string {
	switch e {
	case blocksEncoding:
		return "blocks"
	case sixelEncoding:
		return "sixel"
	case itermEncoding:
		return "iterm"
	case kittyEncoding:
		return "kitty"
	default:
		return "unknown"
	}
}

var desiredEnc int

func init() {
	flag.IntVar(&desiredEnc, "encoding", int(unknownEncoding), "image encoding")

	f, err := os.OpenFile("image.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(f)
}

func main() {
	flag.Parse()

	t := uv.DefaultTerminal()

	// Use altscreen buffer.
	t.EnterAltScreen() //nolint:errcheck

	// Enable mouse support.
	t.WriteString(ansi.SetMode(
		ansi.ModeMouseButtonEvent,
		ansi.ModeMouseExtSgr,
	))

	if err := t.Start(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}

	// Get image info.
	charmImgFile, err := os.Open("./charm.jpg")
	if err != nil {
		log.Fatalf("failed to open image file: %v", err)
	}

	defer charmImgFile.Close() //nolint:errcheck
	charmImgStat, err := charmImgFile.Stat()
	if err != nil {
		log.Fatalf("failed to stat image file: %v", err)
	}

	charmImgFileSize := charmImgStat.Size()

	var charmImgBuf bytes.Buffer
	var charmImgB64 []byte
	imgTee := io.TeeReader(charmImgFile, &charmImgBuf)
	charmImg, _, err := image.Decode(imgTee)
	if err != nil {
		log.Fatalf("failed to decode image: %v", err)
	}

	charmImgArea := charmImg.Bounds()

	// Image related variables.
	var (
		winSize uv.WindowSizeEvent
		pixSize uv.WindowPixelSizeEvent
		imgEnc  = blocksEncoding
	)
	if desiredEnc > 0 {
		imgEnc = imageEncoding(desiredEnc)
	}

	winSize.Width, winSize.Height, err = t.GetSize()
	if err != nil {
		log.Fatalf("failed to get terminal size: %v", err)
	}

	t.Resize(winSize.Width, winSize.Height) //nolint:errcheck

	upgradeEnc := func(enc imageEncoding) {
		if desiredEnc == unknownEncoding {
			if enc > imgEnc {
				imgEnc = enc
			}
		}
	}

	// Check environment variables for supported encodings.
	var (
		termType    = os.Getenv("TERM")
		termProg    string
		lcTerm      string
		termVersion string
		ok          bool
	)
	if termProg, ok = os.LookupEnv("TERM_PROGRAM"); ok {
		if strings.Contains(termProg, "iTerm") ||
			strings.Contains(termProg, "WezTerm") ||
			strings.Contains(termProg, "mintty") ||
			strings.Contains(termProg, "vscode") ||
			strings.Contains(termProg, "Tabby") ||
			strings.Contains(termProg, "Hyper") ||
			strings.Contains(termProg, "rio") {
			upgradeEnc(itermEncoding)
		}
		if lcTerm, ok = os.LookupEnv("LC_TERMINAL"); ok {
			if strings.Contains(lcTerm, "iTerm") {
				upgradeEnc(itermEncoding)
			}
		}
	}

	// Display image methods.
	imgCellSize := func() (int, int) {
		if winSize.Width == 0 || winSize.Height == 0 || pixSize.Width == 0 || pixSize.Height == 0 {
			return 0, 0
		}

		cellW, cellH := pixSize.Width/winSize.Width, pixSize.Height/winSize.Height
		imgW, imgH := charmImgArea.Dx(), charmImgArea.Dy()
		return imgW / cellW, imgH / cellH
	}

	var transmitKitty bool
	var imgCellW, imgCellH int
	var imgOffsetX, imgOffsetY int
	imgCellW, imgCellH = imgCellSize()
	imgOffsetX = winSize.Width/2 - imgCellW/2
	imgOffsetY = winSize.Height/2 - imgCellH/2

	fillStyle := uv.Style{Fg: ansi.IndexedColor(240)}
	displayImg := func() {
		img := charmImg
		imgArea := uv.Rect(
			imgOffsetX,
			imgOffsetY,
			imgCellW,
			imgCellH,
		)
		if !imgArea.In(winSize.Bounds()) {
			imgArea = imgArea.Intersect(winSize.Bounds())
			// TODO: Crop image.
		}

		log.Printf("image area: %v", imgArea)

		// Clear the screen.
		t.Clear()
		fill := uv.Cell{Content: "/", Width: 1, Style: fillStyle}
		t.Fill(&fill)

		// Draw the image on the screen.
		switch imgEnc {
		case blocksEncoding:
			blocks := mosaic.New().Width(imgCellW).Height(imgCellH).Scale(2)
			ss := uv.NewStyledString(blocks.Render(img))
			ss.Draw(t, imgArea)

		case itermEncoding, sixelEncoding:
			t.FillArea(&uv.EmptyCell, imgArea)

		case kittyEncoding:
			const imgId = 31 // random id for kitty graphics
			if !transmitKitty {
				var buf bytes.Buffer
				if err := kitty.EncodeGraphics(&buf, img, &kitty.Options{
					ID:               imgId,
					Action:           kitty.TransmitAndPut,
					Transmission:     kitty.Direct,
					Format:           kitty.RGBA,
					Size:             int(charmImgFileSize),
					ImageWidth:       charmImgArea.Dx(),
					ImageHeight:      charmImgArea.Dy(),
					Columns:          imgArea.Dx(),
					Rows:             imgArea.Dy(),
					VirtualPlacement: true,
					Quite:            2,
				}); err != nil {
					log.Fatalf("failed to encode image for Kitty Graphics: %v", err)
				}

				t.WriteString(buf.String()) //nolint:errcheck
				transmitKitty = true
				t.Flush()
			}

			// Build Kitty graphics unicode place holders
			var fg color.Color
			var extra int
			var r, g, b int
			extra, r, g, b = imgId>>24&0xff, imgId>>16&0xff, imgId>>8&0xff, imgId&0xff

			if r == 0 && g == 0 {
				fg = ansi.IndexedColor(b)
			} else {
				fg = color.RGBA{
					R: uint8(r), //nolint:gosec
					G: uint8(g), //nolint:gosec
					B: uint8(b), //nolint:gosec
					A: 0xff,
				}
			}

			for y := 0; y < imgArea.Dy(); y++ {
				// As an optimization, we only write the fg color sequence id, and
				// column-row data once on the first cell. The terminal will handle
				// the rest.
				content := []rune{kitty.Placeholder, kitty.Diacritic(y), kitty.Diacritic(0)}
				if extra > 0 {
					content = append(content, kitty.Diacritic(extra))
				}
				t.SetCell(imgArea.Min.X, imgArea.Min.Y+y, &uv.Cell{
					Style:   uv.Style{Fg: fg},
					Content: string(content),
					Width:   1,
				})
				for x := 1; x < imgArea.Dx(); x++ {
					t.SetCell(imgArea.Min.X+x, imgArea.Min.Y+y, &uv.Cell{
						Style:   uv.Style{Fg: fg},
						Content: string(kitty.Placeholder),
						Width:   1,
					})
				}
			}

		}

		t.Display() //nolint:errcheck

		switch imgEnc {
		case sixelEncoding:
			var senc sixel.Encoder
			var buf bytes.Buffer
			senc.Encode(&buf, img)
			six := ansi.SixelGraphics(0, 1, 0, buf.Bytes())
			// Note: Sixel starts drawing from the current cursor position
			// and the cursor ends up at the bottom of the image
			// Add a small offset to prevent top cutoff
			if imgArea.Min.Y > 0 {
				t.MoveTo(imgArea.Min.X, imgArea.Min.Y+1)
			} else {
				t.MoveTo(imgArea.Min.X, imgArea.Min.Y)
			}

			t.WriteString(six) //nolint:errcheck
			t.WriteString(ansi.CursorPosition(imgArea.Min.X+1, imgArea.Min.Y+1))

		case itermEncoding:
			// Now, we need to encode the image and place it in the first
			// cell before moving the cursor to the correct position.
			if charmImgB64 == nil {
				// Encode the image to base64 for the first time.
				charmImgB64 = []byte(base64.StdEncoding.EncodeToString(charmImgBuf.Bytes()))
			}

			data := ansi.ITerm2(iterm2.File{
				Name:              "charm.jpg",
				Width:             iterm2.Cells(imgArea.Dx()),
				Height:            iterm2.Cells(imgArea.Dy()),
				Inline:            true,
				Content:           charmImgB64,
				IgnoreAspectRatio: true,
			})

			cup := ansi.CursorPosition(imgArea.Min.X+1, imgArea.Min.Y+1)
			t.MoveTo(imgArea.Min.X, imgArea.Min.Y)
			t.WriteString(data) //nolint:errcheck
			t.WriteString(cup)  //nolint:errcheck
		}

		if t.Buffered() > 0 {
			t.Flush() //nolint:errcheck
		}
	}

	// Query image encoding support.
	t.WriteString(ansi.RequestPrimaryDeviceAttributes)        // Query Sixel support.
	t.WriteString(ansi.RequestNameVersion)                    // Query terminal version and name.
	t.WriteString(ansi.WindowOp(ansi.RequestWindowSizeWinOp)) // Request window size.
	// Query Kitty Graphics support using random id=31.
	t.WriteString(ansi.KittyGraphics([]byte("AAAA"), "i=31", "s=1", "v=1", "a=q", "t=d", "f=24"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for input events.
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case ev := <-t.Events():
			switch ev := ev.(type) {
			case uv.WindowPixelSizeEvent:
				// XXX: This is only emitted with traditional Unix systems. On
				// Windows, we would need to use [ansi.RequestWindowSizeWinOp] to
				// get the pixel size.
				pixSize = ev
				imgCellW, imgCellH = imgCellSize()
				imgOffsetX = winSize.Width/2 - imgCellW/2
				imgOffsetY = winSize.Height/2 - imgCellH/2
				displayImg()
			case uv.WindowSizeEvent:
				winSize = ev
				imgCellW, imgCellH = imgCellSize()
				imgOffsetX = winSize.Width/2 - imgCellW/2
				imgOffsetY = winSize.Height/2 - imgCellH/2
				t.Erase()
				log.Printf("image cell size: %d x %d", imgCellW, imgCellH)
				if err := t.Resize(ev.Width, ev.Height); err != nil {
					log.Fatalf("failed to resize program: %v", err)
				}

				displayImg()
			case uv.KeyPressEvent:
				switch {
				case ev.MatchString("q", "ctrl+c"):
					cancel() // This will stop the loop
				case ev.MatchString("up", "k"):
					imgOffsetY--
				case ev.MatchString("down", "j"):
					imgOffsetY++
				case ev.MatchString("left", "h"):
					imgOffsetX--
				case ev.MatchString("right", "l"):
					imgOffsetX++
				}

				displayImg()
			case uv.MouseClickEvent:
				imgOffsetX = ev.X - (imgCellW / 2)
				imgOffsetY = ev.Y - (imgCellH / 2)

				displayImg()
			case uv.PrimaryDeviceAttributesEvent:
				if slices.Contains(ev, 4) {
					upgradeEnc(sixelEncoding)
					displayImg()
				}

			case uv.TerminalVersionEvent:
				if strings.Contains(ev.Name, "iTerm") || strings.Contains(ev.Name, "WezTerm") {
					upgradeEnc(itermEncoding)
					displayImg()
				}

			case uv.WindowOpEvent:
				// Here 4 corresponds to the window size response.
				if ev.Op == 4 && len(ev.Args) >= 2 {
					pixSize.Height = ev.Args[0]
					pixSize.Width = ev.Args[1]
				}
			case uv.KittyGraphicsEvent:
				if strings.Contains(termType, "wezterm") ||
					strings.Contains(termVersion, "WezTerm") ||
					strings.Contains(termProg, "WezTerm") {
					// WezTerm doesn't support Kitty Unicode Graphics
					break
				}
				if ev.Options.ID == 31 {
					upgradeEnc(kittyEncoding)
				}

				displayImg()
			}
		}
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Disable mouse support.
	t.WriteString(ansi.ResetMode(
		ansi.ModeMouseButtonEvent,
		ansi.ModeMouseExtSgr,
	))

	if err := t.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown program: %v", err)
	}

	fmt.Println("image encoding:", imgEnc)
}

func init() {
	f, err := os.OpenFile("uv_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	log.SetOutput(f)
}
