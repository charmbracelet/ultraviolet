//go:build windows
// +build windows

package uv

import (
	"context"
	"fmt"
	"strings"
	"time"

	xwindows "github.com/charmbracelet/x/windows"
	"github.com/muesli/cancelreader"
	"golang.org/x/sys/windows"
)

type inputRecord = xwindows.InputRecord

// streamData sends data from the input stream to the event channel.
func (d *TerminalReader) streamData(ctx context.Context, readc chan []byte, recordc chan []inputRecord) error {
	cc, ok := d.r.(*conInputReader)
	if !ok {
		d.logf("streamData: reader is not a conInputReader, falling back to default implementation")
		return d.sendBytes(ctx, readc)
	}

	var records []inputRecord
	var err error
	for {
		for {
			records, err = peekNConsoleInputs(cc.conin, readBufSize)
			if cc.isCanceled() {
				return cancelreader.ErrCanceled
			}
			if err != nil {
				return err
			}
			if len(records) > 0 {
				break
			}

			// Sleep for a bit to avoid busy waiting.
			time.Sleep(10 * time.Millisecond)
		}

		records, err = readNConsoleInputs(cc.conin, uint32(len(records))) //nolint:gosec
		if cc.isCanceled() {
			return cancelreader.ErrCanceled
		}
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case recordc <- records:
		}
	}
}

func (d *TerminalReader) processRecords(records []inputRecord, eventc chan<- Event) {
	processWin32Buf := func() {
		// Parse any remaining buffered data and escape sequences.
		b := d.win32Buf.Bytes()
		if n := d.sendEvents(b, true, eventc); n > 0 {
			d.logf("processing buffer, %d bytes, %q", n, b[:n])
			d.win32Buf.Next(n)
		}
	}

	var lastEventType uint16
	for i, record := range records {
		switch record.EventType {
		case xwindows.KEY_EVENT:
			kevent := record.KeyEvent()
			// d.logf("key event: %s", keyEventString(kevent.VirtualKeyCode, kevent.VirtualScanCode, kevent.Char, kevent.KeyDown, kevent.ControlKeyState, kevent.RepeatCount))

			if ev := d.parseWin32InputKeyEvent(kevent.VirtualKeyCode, kevent.VirtualScanCode,
				kevent.Char, kevent.KeyDown, kevent.ControlKeyState, kevent.RepeatCount); ev != nil {
				if d.win32Buf.Len() > 0 {
					processWin32Buf()
				}
				eventc <- ev
			}

		case xwindows.WINDOW_BUFFER_SIZE_EVENT:
			wevent := record.WindowBufferSizeEvent()
			if wevent.Size.X != d.lastWinsizeX || wevent.Size.Y != d.lastWinsizeY {
				d.lastWinsizeX, d.lastWinsizeY = wevent.Size.X, wevent.Size.Y
				eventc <- WindowSizeEvent{
					Width:  int(wevent.Size.X),
					Height: int(wevent.Size.Y),
				}
			}
		case xwindows.MOUSE_EVENT:
			if d.MouseMode == nil || *d.MouseMode == 0 {
				continue
			}
			mouseMode := *d.MouseMode
			mevent := record.MouseEvent()
			event := mouseEvent(d.lastMouseBtns, mevent)
			// We emulate mouse mode levels on Windows. This is because Windows
			// doesn't have a concept of different mouse modes. We use the mouse mode to determine
			switch m := event.(type) {
			case MouseMotionEvent:
				if m.Button == MouseNone && mouseMode&AllMouseMode == 0 {
					continue
				}
				if m.Button != MouseNone && mouseMode&DragMouseMode == 0 {
					continue
				}
			}
			d.lastMouseBtns = mevent.ButtonState
			eventc <- event
		case xwindows.FOCUS_EVENT:
			fevent := record.FocusEvent()
			if fevent.SetFocus {
				eventc <- FocusEvent{}
			} else {
				eventc <- BlurEvent{}
			}
		case xwindows.MENU_EVENT:
			// ignore
		}

		notKeyEvent := lastEventType == xwindows.KEY_EVENT && record.EventType != xwindows.KEY_EVENT
		if d.win32Buf.Len() > 0 && (notKeyEvent || i == len(records)-1) {
			processWin32Buf()
		}

		lastEventType = record.EventType
	}
}

func mouseEventButton(p, s uint32) (MouseButton, bool) {
	var isRelease bool
	button := MouseNone
	btn := p ^ s
	if btn&s == 0 {
		isRelease = true
	}

	if btn == 0 {
		switch {
		case s&xwindows.FROM_LEFT_1ST_BUTTON_PRESSED > 0:
			button = MouseLeft
		case s&xwindows.FROM_LEFT_2ND_BUTTON_PRESSED > 0:
			button = MouseMiddle
		case s&xwindows.RIGHTMOST_BUTTON_PRESSED > 0:
			button = MouseRight
		case s&xwindows.FROM_LEFT_3RD_BUTTON_PRESSED > 0:
			button = MouseBackward
		case s&xwindows.FROM_LEFT_4TH_BUTTON_PRESSED > 0:
			button = MouseForward
		}
		return button, isRelease
	}

	switch btn {
	case xwindows.FROM_LEFT_1ST_BUTTON_PRESSED: // left button
		button = MouseLeft
	case xwindows.RIGHTMOST_BUTTON_PRESSED: // right button
		button = MouseRight
	case xwindows.FROM_LEFT_2ND_BUTTON_PRESSED: // middle button
		button = MouseMiddle
	case xwindows.FROM_LEFT_3RD_BUTTON_PRESSED: // unknown (possibly mouse backward)
		button = MouseBackward
	case xwindows.FROM_LEFT_4TH_BUTTON_PRESSED: // unknown (possibly mouse forward)
		button = MouseForward
	}

	return button, isRelease
}

func mouseEvent(p uint32, e xwindows.MouseEventRecord) (ev Event) {
	var mod KeyMod
	var isRelease bool
	if e.ControlKeyState&(xwindows.LEFT_ALT_PRESSED|xwindows.RIGHT_ALT_PRESSED) != 0 {
		mod |= ModAlt
	}
	if e.ControlKeyState&(xwindows.LEFT_CTRL_PRESSED|xwindows.RIGHT_CTRL_PRESSED) != 0 {
		mod |= ModCtrl
	}
	if e.ControlKeyState&(xwindows.SHIFT_PRESSED) != 0 {
		mod |= ModShift
	}

	m := Mouse{
		X:   int(e.MousePositon.X),
		Y:   int(e.MousePositon.Y),
		Mod: mod,
	}

	wheelDirection := int16(highWord(e.ButtonState)) //nolint:gosec
	switch e.EventFlags {
	case 0, xwindows.DOUBLE_CLICK:
		m.Button, isRelease = mouseEventButton(p, e.ButtonState)
	case xwindows.MOUSE_WHEELED:
		if wheelDirection > 0 {
			m.Button = MouseWheelUp
		} else {
			m.Button = MouseWheelDown
		}
	case xwindows.MOUSE_HWHEELED:
		if wheelDirection > 0 {
			m.Button = MouseWheelRight
		} else {
			m.Button = MouseWheelLeft
		}
	case xwindows.MOUSE_MOVED:
		m.Button, _ = mouseEventButton(p, e.ButtonState)
		return MouseMotionEvent(m)
	}

	if isWheel(m.Button) {
		return MouseWheelEvent(m)
	} else if isRelease {
		return MouseReleaseEvent(m)
	}

	return MouseClickEvent(m)
}

func highWord(data uint32) uint16 {
	return uint16((data & 0xFFFF0000) >> 16) //nolint:gosec
}

func readNConsoleInputs(console windows.Handle, maxEvents uint32) ([]xwindows.InputRecord, error) {
	if maxEvents == 0 {
		return nil, fmt.Errorf("maxEvents cannot be zero")
	}

	records := make([]xwindows.InputRecord, maxEvents)
	n, err := readConsoleInput(console, records)
	return records[:n], err
}

func readConsoleInput(console windows.Handle, inputRecords []xwindows.InputRecord) (uint32, error) {
	if len(inputRecords) == 0 {
		return 0, fmt.Errorf("size of input record buffer cannot be zero")
	}

	var read uint32

	err := xwindows.ReadConsoleInput(console, &inputRecords[0], uint32(len(inputRecords)), &read) //nolint:gosec

	return read, err //nolint:wrapcheck
}

func peekConsoleInput(console windows.Handle, inputRecords []xwindows.InputRecord) (uint32, error) {
	if len(inputRecords) == 0 {
		return 0, fmt.Errorf("size of input record buffer cannot be zero")
	}

	var read uint32

	err := xwindows.PeekConsoleInput(console, &inputRecords[0], uint32(len(inputRecords)), &read) //nolint:gosec

	return read, err //nolint:wrapcheck
}

func peekNConsoleInputs(console windows.Handle, maxEvents uint32) ([]xwindows.InputRecord, error) {
	if maxEvents == 0 {
		return nil, fmt.Errorf("maxEvents cannot be zero")
	}

	records := make([]xwindows.InputRecord, maxEvents)
	n, err := peekConsoleInput(console, records)
	return records[:n], err
}

//nolint:unused
func keyEventString(vkc, sc uint16, r rune, keyDown bool, cks uint32, repeatCount uint16) string {
	var s strings.Builder
	s.WriteString("vkc: ")
	s.WriteString(fmt.Sprintf("%d, 0x%02x", vkc, vkc))
	s.WriteString(", sc: ")
	s.WriteString(fmt.Sprintf("%d, 0x%02x", sc, sc))
	s.WriteString(", r: ")
	s.WriteString(fmt.Sprintf("%q", r))
	s.WriteString(", down: ")
	s.WriteString(fmt.Sprintf("%v", keyDown))
	s.WriteString(", cks: [")
	if cks&xwindows.LEFT_ALT_PRESSED != 0 {
		s.WriteString("left alt, ")
	}
	if cks&xwindows.RIGHT_ALT_PRESSED != 0 {
		s.WriteString("right alt, ")
	}
	if cks&xwindows.LEFT_CTRL_PRESSED != 0 {
		s.WriteString("left ctrl, ")
	}
	if cks&xwindows.RIGHT_CTRL_PRESSED != 0 {
		s.WriteString("right ctrl, ")
	}
	if cks&xwindows.SHIFT_PRESSED != 0 {
		s.WriteString("shift, ")
	}
	if cks&xwindows.CAPSLOCK_ON != 0 {
		s.WriteString("caps lock, ")
	}
	if cks&xwindows.NUMLOCK_ON != 0 {
		s.WriteString("num lock, ")
	}
	if cks&xwindows.SCROLLLOCK_ON != 0 {
		s.WriteString("scroll lock, ")
	}
	if cks&xwindows.ENHANCED_KEY != 0 {
		s.WriteString("enhanced key, ")
	}
	s.WriteString("], repeat count: ")
	s.WriteString(fmt.Sprintf("%d", repeatCount))
	return s.String()
}
