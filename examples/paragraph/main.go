package main

import (
	"context"
	"log"

	"github.com/charmbracelet/uv"
	"github.com/charmbracelet/uv/component/block"
	"github.com/charmbracelet/uv/component/paragraph"
	"github.com/charmbracelet/uv/layout"
	"github.com/charmbracelet/x/ansi"
)

const (
	para1 = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. In quis nibh ipsum. Aliquam a cursus nisl. Mauris dapibus pretium libero, ac aliquam lacus efficitur nec. In sed metus nibh. Suspendisse odio turpis, bibendum a eros eu, commodo euismod felis. Maecenas elementum dui risus, vitae posuere leo aliquet at. Suspendisse potenti. Proin sem odio, congue aliquet sapien eu, aliquam egestas urna. Donec et maximus risus. Nunc diam risus, pretium ut ipsum non, malesuada consequat tellus. Cras non iaculis nunc, eget sollicitudin orci. Ut consequat imperdiet pretium. Fusce commodo, velit nec congue pretium, lectus dui ullamcorper eros, id pretium massa ipsum fringilla eros. Pellentesque dictum, nunc ac rutrum feugiat, eros eros ultrices lacus, finibus finibus arcu enim et nulla.`
	para2 = `Mauris quis aliquet quam. Proin in lectus nulla. Pellentesque pulvinar augue at tortor pharetra ullamcorper. Praesent rutrum lobortis nunc, eget mattis urna commodo eget. Pellentesque lobortis ipsum ex, nec pharetra nulla tempor vel. Praesent maximus ligula a magna vestibulum, eget pulvinar dolor maximus. Quisque ac diam vitae orci fringilla rhoncus at non velit. Nam sed lorem volutpat, ornare nibh in, gravida velit. Nulla condimentum enim tortor, non pulvinar sem feugiat a. In fringilla est vitae ullamcorper eleifend. Mauris non ligula sed neque dignissim porta ac et lorem. Aliquam vitae velit non lacus porttitor lacinia et quis arcu. Ut pharetra scelerisque varius. Morbi nec eros pretium, iaculis quam eu, consectetur dui. Aliquam fringilla ultricies quam. Etiam vel nulla congue, maximus nunc nec, dapibus lacus.`
	para3 = `Nulla ut orci a lacus pharetra feugiat. Sed hendrerit mi sed lectus convallis, viverra posuere leo tristique. Cras eu congue est. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi suscipit blandit enim, id viverra tellus laoreet eget. Curabitur faucibus dapibus ultrices. Nunc sit amet mattis orci. Phasellus ornare nulla a ante auctor elementum sit amet at turpis. Pellentesque vitae urna risus. Donec fringilla ex non bibendum facilisis.`
	para4 = `Nulla sagittis urna id sapien interdum, non semper leo vestibulum. Fusce turpis quam, posuere nec elit id, euismod egestas eros. Nam semper lacus ac purus pharetra consequat. Donec malesuada pretium sollicitudin. Fusce tempor lectus non condimentum interdum. Sed eu mauris eu mi ornare malesuada. Praesent rutrum, lorem et consequat facilisis, nisl tortor fermentum dui, nec tristique lorem dui a lorem. Nulla facilisi. Suspendisse pretium finibus semper.`
	para5 = `Vestibulum sit amet libero accumsan, ultricies leo id, dapibus erat. Phasellus accumsan ante non turpis efficitur, vel dignissim libero finibus. Suspendisse suscipit nulla non neque venenatis lacinia. Vivamus id lorem porta, semper sem non, imperdiet libero. Quisque eu lorem id nisi rhoncus suscipit eu et tortor. In hac habitasse platea dictumst. Etiam venenatis nisi vel ligula sagittis auctor. Maecenas ac nisi elementum, vulputate nibh ut, molestie justo. Praesent quis ligula odio. Donec molestie id metus molestie faucibus. Nunc accumsan vel lectus vitae fringilla. Praesent suscipit venenatis placerat. Cras pellentesque posuere felis eu tempor. Praesent finibus molestie volutpat. Donec varius laoreet quam, eget consectetur leo tristique a. Aenean nec tortor id felis auctor pretium rhoncus at dui.`
)

func main() {
	t := uv.DefaultTerminal()
	if err := t.MakeRaw(); err != nil {
		log.Fatalf("failed to make terminal raw: %v", err)
	}
	if err := t.Start(); err != nil {
		log.Fatalf("failed to start terminal: %v", err)
	}

	t.EnterAltScreen()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create random paragraph areas.
	area1 := uv.Rect(0, 0, 30, 5)
	area2 := uv.Rect(0, 6, 100, 5)
	area3 := uv.Rect(10, 12, 50, 5)
	area4 := uv.Rect(8, 18, 78, 20)
	area5 := uv.Rect(30, 33, 63, 12)

	for ev := range t.Events(ctx) {
		switch ev := ev.(type) {
		case uv.KeyPressEvent:
			cancel()
		case uv.WindowSizeEvent:
			t.Erase()
			t.Resize(ev.Width, ev.Height)
		}

		uv.Clear(t)

		par1 := paragraph.Paragraph{
			Text:  para1,
			Style: uv.Style{Fg: ansi.Green},
		}
		par2 := paragraph.Paragraph{
			Text:  para2,
			Style: uv.Style{Fg: ansi.White, Bg: ansi.Blue},
		}
		par3 := paragraph.Paragraph{
			Text:     para3,
			Style:    uv.Style{Fg: ansi.Red},
			Link:     uv.Link{URL: "https://example.com"},
			Truncate: true,
			Tail:     "...",
		}
		par4 := block.Block{
			Component: paragraph.Paragraph{
				Text:  para4,
				Style: uv.Style{Bg: ansi.Black},
			},
			Style: uv.Style{Bg: ansi.Black},
			Padding: layout.Padding{
				Top:    1,
				Right:  2,
				Bottom: 1,
				Left:   2,
			},
		}
		par5 := paragraph.Paragraph{
			Text:     para5,
			Style:    uv.Style{Fg: ansi.Yellow, Bg: ansi.Magenta},
			Link:     uv.Link{URL: "https://charm.sh"},
			Truncate: true,
			Tail:     " [more]",
		}

		par1.Draw(t, area1)
		par2.Draw(t, area2)
		par3.Draw(t, area3)
		par4.Draw(t, area4)
		par5.Draw(t, area5)

		if err := t.Display(); err != nil {
			log.Fatalf("failed to display terminal: %v", err)
		}
	}

	if err := t.Shutdown(context.Background()); err != nil {
		log.Fatalf("failed to shutdown terminal: %v", err)
	}
}
