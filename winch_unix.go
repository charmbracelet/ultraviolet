package tv

import "github.com/charmbracelet/x/term"

func (l *WinchReceiver) checkSize() (Size, error) {
	w, h, err := term.GetSize(l.out.Fd())
	if err != nil {
		return Size{}, err
	}

	l.lastWidth = w
	l.lastHeight = h

	return Size{w, h}, err
}
