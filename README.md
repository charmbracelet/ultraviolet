# TV

TV, or TeleVision, is a Golang library for building cell-based user interfaces
and programs. It is specifically designed for terminal applications but can
also be used for other types of applications.

## Tutorial

What does a TV consist of? A TV consists of a screen that displays content, has
some sort of input sources that can be used to interact with the screen, and is
meant to display content or programs on the screen.

First, we need to create a screen that will display our program. TV comes with a
`Terminal` screen that is used to display content on a terminal.

```go
t := tv.NewTerminal(os.Stdin, os.Stdout, os.Environ())
// Or simply use...
// t := tv.DefaultTerminal()
```

A terminal screen has a few properties that are unique to it. For example, a
terminal screen can go into raw mode, which is important to disable echoing of
input characters, and to disable signal handling so that we can receive things
like <kbd>ctrl+c</kbd> without the terminal interfering with our program.

Another important property of a terminal screen is the alternate screen buffer.
This property puts the terminal screen into a special mode that allows us to
display content without interfering with the normal screen buffer.

In this tutorial, we will use the alternate screen buffer to display our
program so that we don't affect the normal screen buffer.

```go
// Set the terminal to raw mode.
if err := t.MakeRaw(); err != nil {
  log.Fatal(err)
}

// Make sure we restore the terminal to its original state
// before we exit. We don't care about errors here, but you
// can handle them if you want.
defer t.Restore() //nolint:errcheck

t.EnterAltScreen()
// Make sure we leave the alternate screen buffer
// when we are done with our program.
defer t.LeaveAltScreen()
```

Now that we have our screen set to raw mode and in the alternate screen buffer,
we can create our program that will be displayed on the screen. A program is an
abstraction layer that handles different screen types and implementations. It
only cares about displaying content on the screen.

We need to start our program before we can display anything on the screen. This
will ensure that the program and screen are initialized and ready to display
content. Internally, this will also call `t.Start()` to start the terminal
screen.

```go
p := tv.NewProgram(t)
if err := p.Start(); err != nil {
  log.Fatalf("failed to start program: %v", err)
}
```

Let's display a simple frame with some text in it. A frame is a container that
holds the buffer we're displaying. The final cursor position we want our cursor
to be at, and the viewport area we are working with to display our content.

```go
p.Display(func(f *tv.Frame) error {
  // We will use the StyledString widget to simplify
  // displaying text on the screen.
  // Using [ansi.WcWidth] will ensure that the text is
  // displayed correctly on the screen using traditional
  // terminal width calculations.
  ss := styledstring.New(ansi.WcWidth, "Hello, World!")
  // We want the widget to occupy the given area which
  // is the entire screen because we're using the alternate
  // screen buffer.
  return f.RenderWidget(ss, f.Area)
})
```

Like TVs, different models have different ways to receive input. Some models
have a remote control, while others have a touch screen. A terminal can receive
input from various peripherals usually through control codes and escape
sequences. Our terminal has a `t.Events(ctx)` method that returns a channel
which will receive events from different terminal input sources.

```go
// We want to be able to stop the terminal input loop
// whenever we call cancel().
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

for ev := range t.Events(ctx) {
  switch ev := ev.(type) {
  case tv.WindowSizeEvent:
    // We want to change our program size.
    // Internally, this will also resize our terminal.
    p.Resize(ev.Width, ev.Height)
    // We can also use p.AutoResize() which will
    // query our terminal screen for its size and
    // resize our program to fit the terminal.
    // But, we already have the size from the event
    // so we can use it directly to save a few cycles.
    //p.AutoResize()
  case tv.KeyPressEvent:
    if ev.MatchStrings("q", "ctrl+c") {
      cancel() // This will stop the loop
    }
  }
}
```

Now that we've handled displaying our program and receiving input from the
terminal, we need to handle the program's lifecycle. We need to make sure that
we restore the terminal to its original state when we exit the program. A
program can be stopped gracefully using the `p.Shutdown(ctx)` method.
Internally, this will also call `t.Shutdown(ctx)` to stop the terminal screen.

```go
// We need to make sure we stop the program gracefully
// after we exit the input loop.
if err := p.Shutdown(ctx); err != nil {
  log.Fatal(err)
}
```

Finally, let's put everything together and create a simple program that displays
a frame with "Hello, World!" in it. The program will exit when we press
<kbd>ctrl+c</kbd> or <kbd>q</kbd>.

```go
package main

import (
  "context"
  "log"
  "os"

  "github.com/charmbracelet/tv"
  "github.com/charmbracelet/tv/widget/styledstring"
  "github.com/charmbracelet/x/ansi"
)

func main() {
  // Create a new terminal screen
  t := tv.NewTerminal(os.Stdin, os.Stdout, os.Environ())
  // Or simply use...
  // t := tv.DefaultTerminal()

  // Make sure we restore the terminal to its original state
  // before we exit. We don't care about errors here, but you
  // can handle them if you want.
  defer t.Restore() //nolint:errcheck

  // Enter the alternate screen buffer
  t.EnterAltScreen()
  // Make sure we leave the alternate screen buffer
  // when we are done with our program.
  defer t.LeaveAltScreen()

  // Create a new program
  p := tv.NewProgram(t)
  // Start the program
  if err := p.Start(); err != nil {
    log.Fatalf("failed to start program: %v", err)
  }

  // We want to be able to stop the terminal input loop
  // whenever we call cancel().
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()

  // This will block until we close the events
  // channel or cancel the context.
  for ev := range t.Events(ctx) {
    switch ev := ev.(type) {
    case tv.WindowSizeEvent:
      p.Resize(ev.Width, ev.Height)
    case tv.KeyPressEvent:
      if ev.MatchStrings("q", "ctrl+c") {
        cancel() // This will stop the loop
      }
    }

    // Display the frame with the styled string
    if err := p.Display(func(f *tv.Frame) error {
      // We will use the StyledString widget to simplify
      // displaying text on the screen.
      // Using [ansi.WcWidth] will ensure that the text is
      // displayed correctly on the screen using traditional
      // terminal width calculations.
      ss := styledstring.New(ansi.WcWidth, "Hello, World!")
      // We want the widget to occupy the given area which
      // is the entire screen because we're using the alternate
      // screen buffer.
      return f.RenderWidget(ss, f.Area)
    }); err != nil {
      log.Fatal(err)
    }
  }

  if err := p.Shutdown(ctx); err != nil {
    log.Fatal(err)
  }
}
```

---

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge.jpg" width="400"></a>

Charm热爱开源 • Charm loves open source • نحنُ نحب المصادر المفتوحة
