package main

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/layout"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/charmbracelet/x/ansi"
)

func main() {
	t := uv.DefaultTerminal()
	scr := t.Screen()

	if err := t.Start(); err != nil {
		log.Fatalln("failed to start terminal:", err)
	}

	defer t.Stop()

	var area uv.Rectangle

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			render(scr, area)

			scr.Render()
			scr.Flush()

		case ev := <-t.Events():
			switch ev := ev.(type) {
			case uv.WindowSizeEvent:
				area = ev.Bounds()
				scr.Resize(area.Dx(), area.Dy())
				screen.Clear(scr)

			case uv.KeyPressEvent:
				switch {
				case ev.MatchString("ctrl+c", "q"):
					return
				}
			}
		}
	}
}

func render(sc uv.Screen, area uv.Rectangle) {
	rows := layout.Vertical(
		layout.Len(9),
		layout.Len(9),
		layout.Len(9),
		layout.Len(9),
		layout.Len(9),
		layout.Min(0), // fills remaining space
	).Split(area)

	var areas []uv.Rectangle

	for _, r := range rows {
		cols := layout.Horizontal(
			layout.Len(14),
			layout.Len(14),
			layout.Len(14),
			layout.Len(14),
			layout.Len(14),
			layout.Min(0),
		).Split(r)

		areas = append(areas, cols[:5]...) // ignore Min(0)
	}

	type Named struct {
		Name        string
		Constraints []layout.Constraint
	}

	examples := []Named{
		{
			"Len",
			[]layout.Constraint{
				layout.Len(0),
				layout.Len(2),
				layout.Len(3),
				layout.Len(6),
				layout.Len(10),
				layout.Len(15),
			},
		},
		{
			"Min",
			[]layout.Constraint{
				layout.Min(0),
				layout.Min(2),
				layout.Min(3),
				layout.Min(6),
				layout.Min(10),
				layout.Min(15),
			},
		},
		{
			"Max",
			[]layout.Constraint{
				layout.Max(0),
				layout.Max(2),
				layout.Max(3),
				layout.Max(6),
				layout.Max(10),
				layout.Max(15),
			},
		},
		{
			"Perc",
			[]layout.Constraint{
				layout.Percent(0),
				layout.Percent(25),
				layout.Percent(50),
				layout.Percent(75),
				layout.Percent(100),
				layout.Percent(150),
			},
		},
		{
			"Ratio",
			[]layout.Constraint{
				layout.Ratio{0, 4},
				layout.Ratio{1, 4},
				layout.Ratio{2, 4},
				layout.Ratio{3, 4},
				layout.Ratio{4, 4},
				layout.Ratio{6, 4},
			},
		},
	}

	for i, e := range cartesianProduct(examples, examples) {
		renderExampleCombinations(
			sc,
			areas[i],
			fmt.Sprintf("%s | %s", e.Left.Name, e.Right.Name),
			zip(e.Left.Constraints, e.Right.Constraints),
		)
	}
}

func renderExampleCombinations(
	sc uv.Screen,
	area uv.Rectangle,
	title string,
	constraints []Pair[layout.Constraint, layout.Constraint],
) {
	rows := layout.Vertical(
		slices.Repeat(
			[]layout.Constraint{layout.Len(1)},
			len(constraints)+1,
		)...,
	).
		WithPadding(layout.Pad(1)).
		Split(area)

	screen.NewContext(sc).DrawString(title, rows[0].Min.X, rows[0].Min.Y-1)

	for _, p := range zip(constraints, rows) {
		renderExample(sc, p.Right, p.Left.Left, p.Left.Right, layout.Min(0))
	}

	nums := "123456789012"
	row := rows[len(rows)-1]

	screen.NewContext(sc).DrawString(
		nums[:min(len(nums), row.Dx())],
		row.Min.X,
		row.Min.Y,
	)
}

func renderExample(sc uv.Screen, area uv.Rectangle, constraints ...layout.Constraint) {
	var r, g, b uv.Rectangle

	layout.Horizontal(constraints...).Split(area).Assign(&r, &b, &g)

	screen.FillArea(sc, cell(ansi.Red), r)
	screen.FillArea(sc, cell(ansi.Green), g)
	screen.FillArea(sc, cell(ansi.Blue), b)

	draw := func(bg ansi.Color, s string, r uv.Rectangle) {
		ctx := screen.NewContext(sc).WithStyle(uv.Style{Bg: bg})

		ctx.DrawString(s[:min(len(s), r.Dx())], r.Min.X, r.Min.Y)
	}

	draw(ansi.Red, constraintLabel(constraints[0]), r)
	draw(ansi.Green, strings.Repeat(".", g.Dx()), g)
	draw(ansi.Blue, constraintLabel(constraints[1]), b)
}

func constraintLabel(c layout.Constraint) string {
	switch c := c.(type) {
	case layout.Fill:
		return strconv.Itoa(int(c))

	case layout.Len:
		return strconv.Itoa(int(c))

	case layout.Max:
		return strconv.Itoa(int(c))

	case layout.Min:
		return strconv.Itoa(int(c))

	case layout.Percent:
		return strconv.Itoa(int(c))

	case layout.Ratio:

		return fmt.Sprintf("%d/%d", c.Num, c.Den)
	default:
		return ""
	}
}

func cell(bg ansi.Color) *uv.Cell {
	return &uv.Cell{
		Content: " ",
		Width:   1,
		Style: uv.Style{
			Bg: bg,
		},
	}
}

type Pair[T, U any] struct {
	Left  T
	Right U
}

func zip[T, U any](a []T, b []U) []Pair[T, U] {
	zipped := make([]Pair[T, U], 0, min(len(a), len(b)))

	for i := range min(len(a), len(b)) {
		zipped = append(zipped, Pair[T, U]{Left: a[i], Right: b[i]})
	}

	return zipped
}

func cartesianProduct[T, U any](as []T, bs []U) []Pair[T, U] {
	product := make([]Pair[T, U], 0, len(as)*len(bs))

	for _, a := range as {
		for _, b := range bs {
			product = append(product, Pair[T, U]{Left: a, Right: b})
		}
	}
	return product
}
