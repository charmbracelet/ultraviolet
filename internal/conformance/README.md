# Renderer conformance tests

Fuzz tests that drive the terminal renderer against real terminal emulators and
compare the resulting screen against ground truth.

## Why this is a separate module

The renderer keeps a model of what the terminal is currently showing and writes
only the difference between that model and the next frame. When the model drifts
out of step with reality the output is still valid ANSI and still looks
plausible; it just paints the wrong thing. Asserting on escape sequences cannot
catch that, because there is no single correct sequence to compare against.

Catching it needs a real terminal emulator to read the screen back from, and the
most accurate one available is [libghostty], ghostty's own VT engine. That is a
cgo package requiring Go 1.26 and a C library, and neither cost belongs in the
library that ships to users.

Keeping it in its own module means the root module never sees it: no dependency,
no cgo, no Go version bump. Nothing here reaches anyone who imports ultraviolet.

## Running

Everything needs `libghostty-vt`. Build it once with [zig] and CMake, from this
directory:

```shell
# Build the commit go.mod depends on. go-libghostty's CMakeLists pins the
# ghostty revision it fetches, and a newer one exports a different set of
# symbols than these bindings call, so building its HEAD fails at link time.
version=$(awk '$1 == "go.mitchellh.com/libghostty" { print $2 }' go.mod)
git clone https://github.com/mitchellh/go-libghostty /tmp/go-libghostty
git -C /tmp/go-libghostty checkout "${version##*-}"
make -C /tmp/go-libghostty build
export PKG_CONFIG_PATH="/tmp/go-libghostty/build/_deps/ghostty-src/zig-out/share/pkgconfig"
```

CI does the same thing from `.github/actions/libghostty-vt`, reading the pin out
of the same line of `go.mod`.

Then, still from this directory:

```shell
# Run the fuzz targets over their seed corpus, plus the regular tests.
go test ./...

# Search for new failing inputs. One target at a time.
go test -run='XXX' -fuzz='FuzzRenderer$' -fuzztime=5m
```

Because these tests live outside the root module, `go test ./...` from the
repository root does not run them. They have to be invoked here.

## What is tested

Each target decodes its input as a small program of renderer operations, replays
it, and checks a property of the result. Go's fuzzer mutates the raw bytes and
uses coverage feedback to steer toward programs that reach new renderer code,
which is what finds the combinations nobody thinks to write by hand.

| Target                    | Property |
| ------------------------- | -------- |
| `FuzzRenderer`            | An incrementally rendered screen matches a full repaint of the same frame. |
| `FuzzRendererIdempotent`  | Rendering an unchanged buffer twice leaves the screen alone. |
| `FuzzRedrawResyncs`       | A forced repaint recovers from any state the renderer drifted into. |
| `FuzzScreenShowsContent`  | Every cluster the buffer holds on the last drawn row reaches the screen. |

Each target also runs both ways the renderer can own a screen, chosen per
program:

- **Fullscreen**, where the frame is the terminal and the renderer moves the
  cursor absolutely.
- **Inline**, where the frame is shorter than the terminal and the cursor moves
  relatively. There is no absolute move to fall back on, so a model that loses
  track of the cursor has nothing to recover with, and the rows below the frame
  belong to whatever was on the screen first. The screen is read back in full,
  past the bottom of the frame, because a frame that shrinks has to clear what
  it no longer covers and the abandoned rows are where the residue lands.

  Rows of someone else's output sit above every inline frame, and the screen is
  read back over them too. They are the point rather than scenery: an inline
  renderer finds the top of its frame by counting rows upward from the cursor,
  and a count that overshoots erases into them. With nothing up there, an erase
  that reached a row too far would read back as blanks, which is what blank rows
  look like anyway, and the mistake would not show.

  An inline frame may also collapse to no rows at all, which is the case most
  likely to reach above itself, since there is no last row for the erase below
  the frame to start from.

  Inline screens are never narrowed, only widened. Narrowing makes the terminal
  rewrap the rows it holds, carrying them, and the cursor among them, somewhere
  a relative move cannot find again. What survives that is a property of drawing
  inline rather than a defect, so asserting on it would only produce failures
  nobody can act on.

Every target runs against two emulators, because they disagree about how wide a
grapheme cluster is and that disagreement is the subject of these tests:

- **libghostty** models legacy widths per codepoint, the way stock terminals do,
  and implements DEC mode 2027 so both width models can be exercised.
- **[x/vt]** always measures whole clusters, so it behaves like a terminal with
  mode 2027 enabled.

Neither is redundant. A bug that only appears under legacy widths is invisible
to x/vt, and a disagreement between the two is itself worth investigating.

## The corpus

`testdata/fuzz/` holds the inputs the nightly run has found. Each one was a real
renderer bug when it landed, so keeping it there is what stops the bug coming
back: the `test` job replays the whole corpus on every push, and Go validates it
before any mutation starts.

Add to it rather than pruning it. When a nightly run fails, download the
`crashers-*` artifact and commit the new file alongside the fix.

[libghostty]: https://github.com/mitchellh/go-libghostty
[x/vt]: https://github.com/charmbracelet/x/tree/main/vt
[zig]: https://ghostty.org/docs/install/build
