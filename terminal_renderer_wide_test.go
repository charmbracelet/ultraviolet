package uv

import (
	"bytes"
	"strings"
	"testing"
)

func TestTransformLineWideCJKShift(t *testing.T) {
	cases := []struct {
		name   string
		frames []string
		width  int
		check  []rune
	}{
		{
			name:   "shift CJK right by 1",
			frames: []string{"你好世界", "a你好世界"},
			width:  20,
			check:  []rune{'你', '好', '世', '界'},
		},
		{
			name:   "shift CJK right by 2",
			frames: []string{"你好世界", "ab你好世界"},
			width:  20,
			check:  []rune{'你', '好', '世', '界'},
		},
		{
			name:   "shift CJK left by 1",
			frames: []string{"a你好世界", "你好世界"},
			width:  20,
			check:  []rune{'你', '好', '世', '界'},
		},
		{
			name:   "shift CJK left by 2",
			frames: []string{"ab你好世界", "你好世界"},
			width:  20,
			check:  []rune{'你', '好', '世', '界'},
		},
		{
			name:   "shift back and forth",
			frames: []string{"你好世界", "a你好世界", "你好世界"},
			width:  20,
			check:  []rune{'你', '好', '世', '界'},
		},
		{
			name:   "insert in mixed ASCII/CJK",
			frames: []string{"abc你好def世界ghi", "abcd你好def世界ghi"},
			width:  30,
			check:  []rune{'你', '好', '世', '界'},
		},
		{
			name:   "delete in mixed ASCII/CJK",
			frames: []string{"abcd你好def世界ghi", "abc你好def世界ghi"},
			width:  30,
			check:  []rune{'你', '好', '世', '界'},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			scr := NewScreenBuffer(tc.width, 3)
			r := NewTerminalRenderer(&buf, []string{
				"TERM=xterm-256color",
				"COLORTERM=truecolor",
			})
			area := Rect(0, 0, tc.width, 3)

			for i, text := range tc.frames {
				buf.Reset()
				ss := NewStyledString(text)
				ss.Draw(&scr, area)
				r.Render(scr.RenderBuffer)
				if err := r.Flush(); err != nil {
					t.Fatalf("frame %d: flush failed: %v", i, err)
				}

				if i == 0 {
					continue
				}

				output := buf.String()
				for _, ch := range tc.check {
					if !strings.Contains(output, string(ch)) {
						t.Errorf("frame %d (%q): wide character %q missing from incremental output %q",
							i, text, string(ch), output)
					}
				}
			}
		})
	}
}

// TestTransformLineWideBufferState verifies that after each render the
// internal curbuf exactly matches newbuf for every cell. This catches bugs
// where ANSI output is wrong but the buffer tracking is also wrong
// (hiding the mismatch from simpler output-only tests).
func TestTransformLineWideBufferState(t *testing.T) {
	width, height := 40, 3

	frames := []string{
		"  你好世界",
		"  a你好世界",
		"  你好世界",
		"Hello 你好世界 world",
		"Hello 你好 world",
		"Hello 你好世界测试 world",
		"  Hello 你好世界测试 world",
		"Hello 你好世界测试 world",
		strings.Repeat("你", width/2),
		strings.Repeat("a你", width/3),
		"",
		"你好",
		"ab你好cd世界ef",
		"你好cd世界ef",
		"ab你好世界ef",
	}

	var buf bytes.Buffer
	scr := NewScreenBuffer(width, height)
	r := NewTerminalRenderer(&buf, []string{
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	})
	area := Rect(0, 0, width, height)

	for i, text := range frames {
		buf.Reset()
		ss := NewStyledString(text)
		scr.Fill(nil)
		ss.Draw(&scr, area)
		r.Render(scr.RenderBuffer)
		if err := r.Flush(); err != nil {
			t.Fatalf("frame %d: flush failed: %v", i, err)
		}

		for y := 0; y < height; y++ {
			curLine := r.curbuf.Line(y)
			newLine := scr.RenderBuffer.Line(y)
			for x := 0; x < width; x++ {
				got := curLine.At(x)
				want := newLine.At(x)
				if !cellEqual(got, want) {
					gotStr := "<nil>"
					wantStr := "<nil>"
					if got != nil {
						gotStr = got.Content
					}
					if want != nil {
						wantStr = want.Content
					}
					t.Errorf("frame %d (%q): cell mismatch at (%d,%d): curbuf=%q want=%q",
						i, text, x, y, gotStr, wantStr)
				}
			}
		}
	}
}

// TestTransformLineWideRepeatedChars tests lines where the same wide
// character repeats (would trigger ECH/REP in emitRange for narrow chars).
// Verifies buffer state correctness.
func TestTransformLineWideRepeatedChars(t *testing.T) {
	var buf bytes.Buffer
	width, height := 30, 1
	scr := NewScreenBuffer(width, height)
	r := NewTerminalRenderer(&buf, []string{
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	})
	area := Rect(0, 0, width, height)

	frames := []string{
		"你你你你你",
		"好你你你你你",
		"你你你你你",
		"你你好你你",
	}

	for i, text := range frames {
		buf.Reset()
		ss := NewStyledString(text)
		scr.Fill(nil)
		ss.Draw(&scr, area)
		r.Render(scr.RenderBuffer)
		if err := r.Flush(); err != nil {
			t.Fatalf("frame %d: flush failed: %v", i, err)
		}

		for y := 0; y < height; y++ {
			curLine := r.curbuf.Line(y)
			newLine := scr.RenderBuffer.Line(y)
			for x := 0; x < width; x++ {
				got := curLine.At(x)
				want := newLine.At(x)
				if !cellEqual(got, want) {
					gotStr, wantStr := "<nil>", "<nil>"
					if got != nil {
						gotStr = got.Content
					}
					if want != nil {
						wantStr = want.Content
					}
					t.Errorf("frame %d (%q): cell (%d,%d): got=%q want=%q",
						i, text, x, y, gotStr, wantStr)
				}
			}
		}
	}
}

// TestTransformLineWideLeadingBlanks tests the EL1 (EraseLineLeft) code
// path with wide characters.
func TestTransformLineWideLeadingBlanks(t *testing.T) {
	var buf bytes.Buffer
	width, height := 30, 1
	scr := NewScreenBuffer(width, height)
	r := NewTerminalRenderer(&buf, []string{
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	})
	area := Rect(0, 0, width, height)

	frames := []string{
		"你好世界",
		"    你好世界",
		"        你好世界",
		"你好世界",
		"  你好世界  ",
	}

	for i, text := range frames {
		buf.Reset()
		ss := NewStyledString(text)
		scr.Fill(nil)
		ss.Draw(&scr, area)
		r.Render(scr.RenderBuffer)
		if err := r.Flush(); err != nil {
			t.Fatalf("frame %d: flush failed: %v", i, err)
		}

		for y := 0; y < height; y++ {
			curLine := r.curbuf.Line(y)
			newLine := scr.RenderBuffer.Line(y)
			for x := 0; x < width; x++ {
				got := curLine.At(x)
				want := newLine.At(x)
				if !cellEqual(got, want) {
					gotStr, wantStr := "<nil>", "<nil>"
					if got != nil {
						gotStr = got.Content
					}
					if want != nil {
						wantStr = want.Content
					}
					t.Errorf("frame %d (%q): cell (%d,%d): got=%q want=%q",
						i, text, x, y, gotStr, wantStr)
				}
			}
		}
	}
}

// TestTransformLineWideMultiLine exercises cursor movement across multiple
// lines containing wide characters. The relativeCursorMove overwrite
// optimization must not emit spaces for zero-width placeholders.
func TestTransformLineWideMultiLine(t *testing.T) {
	width, height := 30, 5

	// Frame pairs: each inner slice is drawn as consecutive frames.
	// Line 0 changes between frames; lines 1-2 have wide chars that
	// the cursor must traverse without corruption.
	type frame struct {
		lines []string
	}
	scenarios := []struct {
		name   string
		frames []frame
	}{
		{
			name: "change line 0 with CJK on lines below",
			frames: []frame{
				{lines: []string{"hello world", "你好世界", "测试数据"}},
				{lines: []string{"HELLO WORLD", "你好世界", "测试数据"}},
				{lines: []string{"hello world", "你好世界", "测试数据"}},
			},
		},
		{
			name: "change last line with CJK on lines above",
			frames: []frame{
				{lines: []string{"你好世界", "测试数据", "aaa"}},
				{lines: []string{"你好世界", "测试数据", "bbb"}},
			},
		},
		{
			name: "change multiple lines with wide chars",
			frames: []frame{
				{lines: []string{"AAAA", "你好世界", "BBBB", "测试数据", "CCCC"}},
				{lines: []string{"aaaa", "你好世界X", "bbbb", "测试X数据", "cccc"}},
			},
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			var buf bytes.Buffer
			scr := NewScreenBuffer(width, height)
			r := NewTerminalRenderer(&buf, []string{
				"TERM=xterm-256color",
				"COLORTERM=truecolor",
			})

			for fi, fr := range sc.frames {
				buf.Reset()
				scr.Fill(nil)
				for li, line := range fr.lines {
					if li >= height {
						break
					}
					area := Rect(0, li, width, li+1)
					ss := NewStyledString(line)
					ss.Draw(&scr, area)
				}
				r.Render(scr.RenderBuffer)
				if err := r.Flush(); err != nil {
					t.Fatalf("frame %d: flush failed: %v", fi, err)
				}

				for y := 0; y < height; y++ {
					curLine := r.curbuf.Line(y)
					newLine := scr.RenderBuffer.Line(y)
					for x := 0; x < width; x++ {
						got := curLine.At(x)
						want := newLine.At(x)
						if !cellEqual(got, want) {
							gotStr, wantStr := "<nil>", "<nil>"
							if got != nil {
								gotStr = got.Content
							}
							if want != nil {
								wantStr = want.Content
							}
							t.Errorf("frame %d: cell (%d,%d): got=%q want=%q",
								fi, x, y, gotStr, wantStr)
						}
					}
				}
			}
		})
	}
}

func TestTransformLineWideCJKStreaming(t *testing.T) {
	var buf bytes.Buffer
	scr := NewScreenBuffer(30, 3)
	r := NewTerminalRenderer(&buf, []string{
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	})
	area := Rect(0, 0, 30, 3)

	texts := []string{
		"你",
		"你好",
		"你好世",
		"你好世界",
		"你好世界，",
		"你好世界，这",
		"你好世界，这是",
		"你好世界，这是一",
		"你好世界，这是一个",
		"你好世界，这是一个测",
		"你好世界，这是一个测试",
	}

	for i, text := range texts {
		buf.Reset()
		ss := NewStyledString(text)
		ss.Draw(&scr, area)
		r.Render(scr.RenderBuffer)
		if err := r.Flush(); err != nil {
			t.Fatalf("step %d: flush failed: %v", i, err)
		}

		if i == 0 {
			continue
		}

		output := buf.String()
		runes := []rune(text)
		lastChar := string(runes[len(runes)-1])
		if !strings.Contains(output, lastChar) {
			t.Errorf("step %d (%q): new character %q missing from incremental output %q",
				i, text, lastChar, output)
		}
	}
}
