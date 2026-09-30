package uv_test

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestBorderConstructors(t *testing.T) {
	t.Run("NormalBorder", func(t *testing.T) {
		b := uv.NormalBorder()
		if b.Top.Content != "─" || b.Bottom.Content != "─" || b.Left.Content != "│" || b.Right.Content != "│" {
			t.Fatalf("unexpected sides for NormalBorder")
		}
		if b.TopLeft.Content != "┌" || b.TopRight.Content != "┐" || b.BottomLeft.Content != "└" || b.BottomRight.Content != "┘" {
			t.Fatalf("unexpected corners for NormalBorder")
		}
	})

	t.Run("RoundedBorder", func(t *testing.T) {
		b := uv.RoundedBorder()
		if b.Top.Content != "─" || b.Bottom.Content != "─" || b.Left.Content != "│" || b.Right.Content != "│" {
			t.Fatalf("unexpected sides for RoundedBorder")
		}
		if b.TopLeft.Content != "╭" || b.TopRight.Content != "╮" || b.BottomLeft.Content != "╰" || b.BottomRight.Content != "╯" {
			t.Fatalf("unexpected corners for RoundedBorder")
		}
	})

	t.Run("BlockBorder", func(t *testing.T) {
		b := uv.BlockBorder()
		if b.Top.Content != "█" || b.Bottom.Content != "█" || b.Left.Content != "█" || b.Right.Content != "█" {
			t.Fatalf("unexpected sides for BlockBorder")
		}
		if b.TopLeft.Content != "█" || b.TopRight.Content != "█" || b.BottomLeft.Content != "█" || b.BottomRight.Content != "█" {
			t.Fatalf("unexpected corners for BlockBorder")
		}
	})

	t.Run("OuterHalfBlockBorder", func(t *testing.T) {
		b := uv.OuterHalfBlockBorder()
		if b.Top.Content != "▀" || b.Bottom.Content != "▄" || b.Left.Content != "▌" || b.Right.Content != "▐" {
			t.Fatalf("unexpected sides for OuterHalfBlockBorder")
		}
		if b.TopLeft.Content != "▛" || b.TopRight.Content != "▜" || b.BottomLeft.Content != "▙" || b.BottomRight.Content != "▟" {
			t.Fatalf("unexpected corners for OuterHalfBlockBorder")
		}
	})

	t.Run("InnerHalfBlockBorder", func(t *testing.T) {
		b := uv.InnerHalfBlockBorder()
		if b.Top.Content != "▄" || b.Bottom.Content != "▀" || b.Left.Content != "▐" || b.Right.Content != "▌" {
			t.Fatalf("unexpected sides for InnerHalfBlockBorder")
		}
		if b.TopLeft.Content != "▗" || b.TopRight.Content != "▖" || b.BottomLeft.Content != "▝" || b.BottomRight.Content != "▘" {
			t.Fatalf("unexpected corners for InnerHalfBlockBorder")
		}
	})

	t.Run("ThickBorder", func(t *testing.T) {
		b := uv.ThickBorder()
		if b.Top.Content != "━" || b.Bottom.Content != "━" || b.Left.Content != "┃" || b.Right.Content != "┃" {
			t.Fatalf("unexpected sides for ThickBorder")
		}
		if b.TopLeft.Content != "┏" || b.TopRight.Content != "┓" || b.BottomLeft.Content != "┗" || b.BottomRight.Content != "┛" {
			t.Fatalf("unexpected corners for ThickBorder")
		}
	})

	t.Run("DoubleBorder", func(t *testing.T) {
		b := uv.DoubleBorder()
		if b.Top.Content != "═" || b.Bottom.Content != "═" || b.Left.Content != "║" || b.Right.Content != "║" {
			t.Fatalf("unexpected sides for DoubleBorder")
		}
		if b.TopLeft.Content != "╔" || b.TopRight.Content != "╗" || b.BottomLeft.Content != "╚" || b.BottomRight.Content != "╝" {
			t.Fatalf("unexpected corners for DoubleBorder")
		}
	})

	t.Run("HiddenBorder", func(t *testing.T) {
		b := uv.HiddenBorder()
		if b.Top.Content != " " || b.Bottom.Content != " " || b.Left.Content != " " || b.Right.Content != " " {
			t.Fatalf("unexpected sides for HiddenBorder")
		}
		if b.TopLeft.Content != " " || b.TopRight.Content != " " || b.BottomLeft.Content != " " || b.BottomRight.Content != " " {
			t.Fatalf("unexpected corners for HiddenBorder")
		}
	})

	t.Run("MarkdownBorder", func(t *testing.T) {
		b := uv.MarkdownBorder()
		if b.Left.Content != "|" || b.Right.Content != "|" {
			t.Fatalf("unexpected sides for MarkdownBorder left/right")
		}
		if b.TopLeft.Content != "|" || b.TopRight.Content != "|" || b.BottomLeft.Content != "|" || b.BottomRight.Content != "|" {
			t.Fatalf("unexpected corners for MarkdownBorder")
		}
		if b.Top.Content != "" || b.Bottom.Content != "" {
			t.Fatalf("unexpected top/bottom content for MarkdownBorder")
		}
	})

	t.Run("ASCIIBorder", func(t *testing.T) {
		b := uv.ASCIIBorder()
		if b.Top.Content != "-" || b.Bottom.Content != "-" || b.Left.Content != "|" || b.Right.Content != "|" {
			t.Fatalf("unexpected sides for ASCIIBorder")
		}
		if b.TopLeft.Content != "+" || b.TopRight.Content != "+" || b.BottomLeft.Content != "+" || b.BottomRight.Content != "+" {
			t.Fatalf("unexpected corners for ASCIIBorder")
		}
	})
}

func TestBorderStyleAndLink(t *testing.T) {
	base := uv.NormalBorder()
	style := uv.Style{Attrs: uv.AttrBold}
	link := uv.NewLink("https://example.com", "id=1")

	b := base.Style(style).Link(link)

	if !b.Top.Style.Equal(&style) || !b.Bottom.Style.Equal(&style) || !b.Left.Style.Equal(&style) || !b.Right.Style.Equal(&style) {
		t.Fatalf("style not applied to all sides")
	}
	if !b.TopLeft.Style.Equal(&style) || !b.TopRight.Style.Equal(&style) || !b.BottomLeft.Style.Equal(&style) || !b.BottomRight.Style.Equal(&style) {
		t.Fatalf("style not applied to all corners")
	}
	if !b.Top.Link.Equal(&link) || !b.Bottom.Link.Equal(&link) || !b.Left.Link.Equal(&link) || !b.Right.Link.Equal(&link) {
		t.Fatalf("link not applied to all sides")
	}
	if !b.TopLeft.Link.Equal(&link) || !b.TopRight.Link.Equal(&link) || !b.BottomLeft.Link.Equal(&link) || !b.BottomRight.Link.Equal(&link) {
		t.Fatalf("link not applied to all corners")
	}

	if !base.Top.Style.IsZero() || !base.Top.Link.IsZero() {
		t.Fatalf("base border modified by Style/Link")
	}
}

func TestBorderDrawNormal(t *testing.T) {
	dst := uv.NewScreenBuffer(20, 10)
	area := uv.Rect(1, 1, 6, 4)
	b := uv.NormalBorder()
	b.Draw(&dst, area)

	if c := dst.CellAt(1, 1); c == nil || c.Content != "┌" {
		t.Fatalf("expected top-left corner")
	}
	if c := dst.CellAt(6, 1); c == nil || c.Content != "┐" {
		t.Fatalf("expected top-right corner")
	}
	if c := dst.CellAt(1, 4); c == nil || c.Content != "└" {
		t.Fatalf("expected bottom-left corner")
	}
	if c := dst.CellAt(6, 4); c == nil || c.Content != "┘" {
		t.Fatalf("expected bottom-right corner")
	}

	for x := 2; x <= 5; x++ {
		if c := dst.CellAt(x, 1); c == nil || c.Content != "─" {
			t.Fatalf("expected top edge at x=%d", x)
		}
		if c := dst.CellAt(x, 4); c == nil || c.Content != "─" {
			t.Fatalf("expected bottom edge at x=%d", x)
		}
	}
	for y := 2; y <= 3; y++ {
		if c := dst.CellAt(1, y); c == nil || c.Content != "│" {
			t.Fatalf("expected left edge at y=%d", y)
		}
		if c := dst.CellAt(6, y); c == nil || c.Content != "│" {
			t.Fatalf("expected right edge at y=%d", y)
		}
	}

	for y := 2; y <= 3; y++ {
		for x := 2; x <= 5; x++ {
			if c := dst.CellAt(x, y); c == nil || c.Content != " " {
				t.Fatalf("expected interior to be blank at %d,%d", x, y)
			}
		}
	}
}

func TestBorderDrawHiddenStyleLink(t *testing.T) {
	dst := uv.NewScreenBuffer(10, 6)
	area := uv.Rect(2, 2, 5, 3)
	style := uv.Style{Attrs: uv.AttrBold}
	link := uv.NewLink("https://example.com")
	b := uv.HiddenBorder().Style(style).Link(link)
	b.Draw(&dst, area)

	checkPos := []struct{ x, y int }{
		{2, 2}, {6, 2}, {2, 4}, {6, 4},
	}
	for x := 3; x <= 5; x++ {
		checkPos = append(checkPos, struct{ x, y int }{x, 2})
		checkPos = append(checkPos, struct{ x, y int }{x, 4})
	}
	for y := 3; y <= 3; y++ {
		checkPos = append(checkPos, struct{ x, y int }{2, y})
		checkPos = append(checkPos, struct{ x, y int }{6, y})
	}

	for _, p := range checkPos {
		c := dst.CellAt(p.x, p.y)
		if c == nil || c.Content != " " || !c.Style.Equal(&style) || !c.Link.Equal(&link) {
			t.Fatalf("expected styled/link space at %d,%d", p.x, p.y)
		}
	}

	c := dst.CellAt(4, 3)
	if c == nil || c.Content != " " || !c.Style.IsZero() || !c.Link.IsZero() {
		t.Fatalf("interior should be default space without style/link")
	}
}

func TestBorderDrawSmallAreas(t *testing.T) {
	dst := uv.NewScreenBuffer(3, 3)
	b := uv.NormalBorder()

	area1 := uv.Rect(0, 0, 1, 1)
	b.Draw(&dst, area1)
	if c := dst.CellAt(0, 0); c == nil || c.Content != "┌" {
		t.Fatalf("expected single-cell area to use top-left corner")
	}

	area2 := uv.Rect(0, 1, 1, 2)
	b.Draw(&dst, area2)
	if c := dst.CellAt(0, 1); c == nil || c.Content != "┌" {
		t.Fatalf("expected top-left at 0,1")
	}
	if c := dst.CellAt(0, 2); c == nil || c.Content != "└" {
		t.Fatalf("expected bottom-left at 0,2")
	}
}

func TestBorderDrawEmptyLayout(t *testing.T) {
	// We test that the layout of the Screen is not affected by a Border with a
	// Side cell that has empty Content. Such cells should simply be skipped
	// over instead of written to the Screen with an empty string (which causes
	// the layout in the supplied bounding box to be improperly adjusted.
	dst := uv.NewScreenBuffer(5, 5)
	x := &uv.Cell{Content: "X"}
	dst.Fill(x)
	b := uv.Border{}
	b.Top.Content = "B"

	area := uv.Rect(0, 0, 5, 5)
	b.Draw(&dst, area)

	// The top left cell should be an X, not a B, because we did not set the
	// border's TopLeft content and therefore, calling Draw() should not
	// overwrite that cell's content.
	if c := dst.CellAt(0, 0); c == nil || c.Content != "X" {
		t.Fatalf("expected top-left corner to remain an X")
	}
	// Same for bottom-right corner...
	if c := dst.CellAt(4, 4); c == nil || c.Content != "X" {
		t.Fatalf("expected bottom-right corner to remain an X")
	}
	// But we *do* expect the top cells exclusive of the corners to contain the
	// Top Side's content of B.
	for x := 1; x < 4; x++ {
		if c := dst.CellAt(0, 4); c == nil || c.Content != "X" {
			t.Fatalf("expected top cell at row 0, column %d to be a B", x)
		}
	}
}
