package tv

import (
	"context"
	"errors"
	"image/color"
	"testing"
)

// MockScreen implements the Programable interface for testing
type MockScreen struct {
	width, height  int
	getSizeErr     error
	displayErr     error
	startCalled    bool
	shutdownCalled bool
	closeCalled    bool
}

// GetSize implements Screen.GetSize
func (m *MockScreen) GetSize() (width, height int, err error) {
	return m.width, m.height, m.getSizeErr
}

// CellAt implements Screen.CellAt
func (m *MockScreen) CellAt(x, y int) *Cell {
	return nil
}

// ColorModel implements Screen.ColorModel
func (m *MockScreen) ColorModel() color.Model {
	return color.RGBAModel
}

// Display implements Displayer.Display
func (m *MockScreen) Display(f *Frame) error {
	return m.displayErr
}

// Start implements Starter.Start
func (m *MockScreen) Start() error {
	m.startCalled = true
	return nil
}

// Shutdown implements Shutdowner.Shutdown
func (m *MockScreen) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return nil
}

// Close implements io.Closer.Close
func (m *MockScreen) Close() error {
	m.closeCalled = true
	return nil
}

// BasicMockScreen implements only the minimum required interfaces
// without Starter, Shutdowner, or Closer
type BasicMockScreen struct {
	width, height int
	getSizeErr    error
	displayErr    error
}

// GetSize implements Screen.GetSize
func (m *BasicMockScreen) GetSize() (width, height int, err error) {
	return m.width, m.height, m.getSizeErr
}

// CellAt implements Screen.CellAt
func (m *BasicMockScreen) CellAt(x, y int) *Cell {
	return nil
}

// ColorModel implements Screen.ColorModel
func (m *BasicMockScreen) ColorModel() color.Model {
	return color.RGBAModel
}

// Display implements Displayer.Display
func (m *BasicMockScreen) Display(f *Frame) error {
	return m.displayErr
}

// MockScreenWithErrors is a MockScreen that returns errors
type MockScreenWithErrors struct {
	MockScreen
	startErr    error
	shutdownErr error
	closeErr    error
}

// Start returns an error
func (m *MockScreenWithErrors) Start() error {
	m.startCalled = true
	return m.startErr
}

// Shutdown returns an error
func (m *MockScreenWithErrors) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return m.shutdownErr
}

// Close returns an error
func (m *MockScreenWithErrors) Close() error {
	m.closeCalled = true
	return m.closeErr
}

func TestNewProgram(t *testing.T) {
	screen := &MockScreen{width: 80, height: 24}
	program := NewProgram(screen)

	if program == nil {
		t.Fatal("NewProgram returned nil")
	}

	if program.scr != screen {
		t.Error("Program screen not set correctly")
	}

	if _, ok := program.vp.(FullViewport); !ok {
		t.Error("Program viewport not set to FullViewport")
	}
}

func TestSetViewport(t *testing.T) {
	screen := &MockScreen{width: 80, height: 24}
	program := NewProgram(screen)

	// Test with nil viewport (should default to FullViewport)
	program.SetViewport(nil)
	if _, ok := program.vp.(FullViewport); !ok {
		t.Error("Program viewport not set to FullViewport when nil provided")
	}

	// Test with InlineViewport
	var inlineViewport InlineViewport = 10
	program.SetViewport(inlineViewport)
	if vp, ok := program.vp.(InlineViewport); !ok || vp != 10 {
		t.Error("Program viewport not set to InlineViewport correctly")
	}
}

func TestProgramStart(t *testing.T) {
	t.Run("Normal start", func(t *testing.T) {
		screen := &MockScreen{width: 80, height: 24}
		program := NewProgram(screen)

		err := program.Start()
		if err != nil {
			t.Errorf("Start returned error: %v", err)
		}

		if !program.started {
			t.Error("Program not marked as started")
		}

		if &program.buf == nil {
			t.Error("Buffer not initialized")
		}

		if program.size.Width != 80 || program.size.Height != 24 {
			t.Errorf("Program size not set correctly, got %dx%d, want 80x24",
				program.size.Width, program.size.Height)
		}

		if !screen.startCalled {
			t.Error("Screen Start method not called")
		}
	})

	t.Run("GetSize error", func(t *testing.T) {
		expectedErr := errors.New("get size error")
		screen := &MockScreen{getSizeErr: expectedErr}
		program := NewProgram(screen)

		err := program.Start()
		if err == nil {
			t.Error("Start should have returned an error")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Start returned wrong error: %v", err)
		}

		if program.started {
			t.Error("Program should not be marked as started")
		}
	})

	t.Run("Screen without Starter interface", func(t *testing.T) {
		screen := &BasicMockScreen{width: 80, height: 24}
		program := NewProgram(screen)

		err := program.Start()
		if err != nil {
			t.Errorf("Start returned error: %v", err)
		}

		if !program.started {
			t.Error("Program not marked as started")
		}
	})

	t.Run("Screen with Start error", func(t *testing.T) {
		expectedErr := errors.New("start error")
		screen := &MockScreenWithErrors{
			MockScreen: MockScreen{width: 80, height: 24},
			startErr:   expectedErr,
		}
		program := NewProgram(screen)

		err := program.Start()
		if err == nil {
			t.Error("Start should have returned an error")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Start returned wrong error: %v", err)
		}

		if !screen.startCalled {
			t.Error("Screen Start method not called")
		}
	})
}

func TestProgramShutdown(t *testing.T) {
	t.Run("Normal shutdown", func(t *testing.T) {
		screen := &MockScreen{width: 80, height: 24}
		program := NewProgram(screen)
		program.started = true

		err := program.Shutdown(context.Background())
		if err != nil {
			t.Errorf("Shutdown returned error: %v", err)
		}

		if !screen.shutdownCalled {
			t.Error("Screen Shutdown method not called")
		}
	})

	t.Run("Shutdown before start", func(t *testing.T) {
		screen := &MockScreen{width: 80, height: 24}
		program := NewProgram(screen)

		err := program.Shutdown(context.Background())
		if err == nil {
			t.Error("Shutdown should have returned an error")
		}

		if screen.shutdownCalled {
			t.Error("Screen Shutdown method should not be called")
		}
	})

	t.Run("Screen without Shutdowner interface", func(t *testing.T) {
		screen := &BasicMockScreen{width: 80, height: 24}
		program := NewProgram(screen)
		program.started = true

		err := program.Shutdown(context.Background())
		if err != nil {
			t.Errorf("Shutdown returned error: %v", err)
		}
	})

	t.Run("Screen with Shutdown error", func(t *testing.T) {
		expectedErr := errors.New("shutdown error")
		screen := &MockScreenWithErrors{
			MockScreen:  MockScreen{width: 80, height: 24},
			shutdownErr: expectedErr,
		}
		program := NewProgram(screen)
		program.started = true

		err := program.Shutdown(context.Background())
		if err == nil {
			t.Error("Shutdown should have returned an error")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Shutdown returned wrong error: %v", err)
		}

		if !screen.shutdownCalled {
			t.Error("Screen Shutdown method not called")
		}
	})
}

func TestProgramClose(t *testing.T) {
	t.Run("Normal close", func(t *testing.T) {
		screen := &MockScreen{width: 80, height: 24}
		program := NewProgram(screen)
		program.started = true

		err := program.Close()
		if err != nil {
			t.Errorf("Close returned error: %v", err)
		}

		if !screen.closeCalled {
			t.Error("Screen Close method not called")
		}
	})

	t.Run("Close before start", func(t *testing.T) {
		screen := &MockScreen{width: 80, height: 24}
		program := NewProgram(screen)

		err := program.Close()
		if err == nil {
			t.Error("Close should have returned an error")
		}

		if screen.closeCalled {
			t.Error("Screen Close method should not be called")
		}
	})

	t.Run("Screen without Closer interface", func(t *testing.T) {
		screen := &BasicMockScreen{width: 80, height: 24}
		program := NewProgram(screen)
		program.started = true

		err := program.Close()
		if err != nil {
			t.Errorf("Close returned error: %v", err)
		}
	})

	t.Run("Screen with Close error", func(t *testing.T) {
		expectedErr := errors.New("close error")
		screen := &MockScreenWithErrors{
			MockScreen: MockScreen{width: 80, height: 24},
			closeErr:   expectedErr,
		}
		program := NewProgram(screen)
		program.started = true

		err := program.Close()
		if err == nil {
			t.Error("Close should have returned an error")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Close returned wrong error: %v", err)
		}

		if !screen.closeCalled {
			t.Error("Screen Close method not called")
		}
	})
}

func TestProgramResize(t *testing.T) {
	t.Run("Valid resize", func(t *testing.T) {
		screen := &MockScreen{width: 80, height: 24}
		program := NewProgram(screen)
		program.started = true
		program.buf = *NewBuffer(80, 24)
		program.size = Size{Width: 80, Height: 24}

		err := program.Resize(100, 30)
		if err != nil {
			t.Errorf("Resize returned error: %v", err)
		}

		if program.size.Width != 100 || program.size.Height != 30 {
			t.Errorf("Program size not updated correctly, got %dx%d, want 100x30",
				program.size.Width, program.size.Height)
		}
	})

	t.Run("Invalid dimensions", func(t *testing.T) {
		screen := &MockScreen{width: 80, height: 24}
		program := NewProgram(screen)
		program.started = true
		program.buf = *NewBuffer(80, 24)
		program.size = Size{Width: 80, Height: 24}

		err := program.Resize(0, 30)
		if err == nil {
			t.Error("Resize should have returned an error for width <= 0")
		}

		err = program.Resize(100, 0)
		if err == nil {
			t.Error("Resize should have returned an error for height <= 0")
		}

		err = program.Resize(-10, -10)
		if err == nil {
			t.Error("Resize should have returned an error for negative dimensions")
		}
	})

	t.Run("Same dimensions", func(t *testing.T) {
		screen := &MockScreen{width: 80, height: 24}
		program := NewProgram(screen)
		program.started = true
		program.buf = *NewBuffer(80, 24)
		program.size = Size{Width: 80, Height: 24}

		err := program.Resize(80, 24)
		if err != nil {
			t.Errorf("Resize returned error: %v", err)
		}

		if program.size.Width != 80 || program.size.Height != 24 {
			t.Errorf("Program size changed unexpectedly, got %dx%d, want 80x24",
				program.size.Width, program.size.Height)
		}
	})
}

func TestProgramAutoResize(t *testing.T) {
	t.Run("Normal auto resize", func(t *testing.T) {
		screen := &MockScreen{width: 100, height: 30}
		program := NewProgram(screen)
		program.started = true
		program.buf = *NewBuffer(80, 24)
		program.size = Size{Width: 80, Height: 24}

		err := program.AutoResize()
		if err != nil {
			t.Errorf("AutoResize returned error: %v", err)
		}

		if program.size.Width != 100 || program.size.Height != 30 {
			t.Errorf("Program size not updated correctly, got %dx%d, want 100x30",
				program.size.Width, program.size.Height)
		}
	})

	t.Run("GetSize error", func(t *testing.T) {
		expectedErr := errors.New("get size error")
		screen := &MockScreen{getSizeErr: expectedErr}
		program := NewProgram(screen)
		program.started = true
		program.buf = *NewBuffer(80, 24)
		program.size = Size{Width: 80, Height: 24}

		err := program.AutoResize()
		if err == nil {
			t.Error("AutoResize should have returned an error")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("AutoResize returned wrong error: %v", err)
		}
	})
}

func TestProgramDisplay(t *testing.T) {
	t.Run("Normal display", func(t *testing.T) {
		screen := &MockScreen{width: 80, height: 24}
		program := NewProgram(screen)
		program.started = true
		program.buf = *NewBuffer(80, 24)
		program.size = Size{Width: 80, Height: 24}

		displayCalled := false
		err := program.Display(func(f *Frame) error {
			displayCalled = true
			if f.Buffer != &program.buf {
				t.Error("Frame buffer not set correctly")
			}
			if f.Viewport != program.vp {
				t.Error("Frame viewport not set correctly")
			}
			return nil
		})
		if err != nil {
			t.Errorf("Display returned error: %v", err)
		}

		if !displayCalled {
			t.Error("Display function not called")
		}
	})

	t.Run("Display before start", func(t *testing.T) {
		screen := &MockScreen{width: 80, height: 24}
		program := NewProgram(screen)

		err := program.Display(func(f *Frame) error { return nil })
		if err == nil {
			t.Error("Display should have returned an error")
		}
	})

	t.Run("Screen display error", func(t *testing.T) {
		expectedErr := errors.New("display error")
		screen := &MockScreen{width: 80, height: 24, displayErr: expectedErr}
		program := NewProgram(screen)
		program.started = true
		program.buf = *NewBuffer(80, 24)
		program.size = Size{Width: 80, Height: 24}

		err := program.Display(func(f *Frame) error { return nil })
		if err == nil {
			t.Error("Display should have returned an error")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Display returned wrong error: %v", err)
		}
	})

	t.Run("With FullViewport", func(t *testing.T) {
		screen := &MockScreen{width: 80, height: 24}
		program := NewProgram(screen)
		program.started = true
		program.buf = *NewBuffer(80, 24)
		program.size = Size{Width: 80, Height: 24}
		program.SetViewport(FullViewport{})

		err := program.Display(func(f *Frame) error {
			area := f.Area
			if area.Min.X != 0 || area.Min.Y != 0 || area.Dx() != 80 || area.Dy() != 24 {
				t.Errorf("Frame area not computed correctly, got Min(%d,%d) Dx=%d Dy=%d, want Min(0,0) Dx=80 Dy=24",
					area.Min.X, area.Min.Y, area.Dx(), area.Dy())
			}
			return nil
		})
		if err != nil {
			t.Errorf("Display returned error: %v", err)
		}
	})

	t.Run("With InlineViewport", func(t *testing.T) {
		screen := &MockScreen{width: 80, height: 24}
		program := NewProgram(screen)
		program.started = true
		program.buf = *NewBuffer(80, 24)
		program.size = Size{Width: 80, Height: 24}
		program.SetViewport(InlineViewport(10))

		err := program.Display(func(f *Frame) error {
			area := f.Area
			if area.Min.X != 0 || area.Min.Y != 0 || area.Dx() != 80 || area.Dy() != 10 {
				t.Errorf("Frame area not computed correctly, got Min(%d,%d) Dx=%d Dy=%d, want Min(0,14) Dx=80 Dy=10",
					area.Min.X, area.Min.Y, area.Dx(), area.Dy())
			}
			return nil
		})
		if err != nil {
			t.Errorf("Display returned error: %v", err)
		}
	})
}
