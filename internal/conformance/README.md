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

Everything needs `libghostty-vt`. Build it once with [zig] and CMake:

```shell
git clone https://github.com/mitchellh/go-libghostty
cd go-libghostty && make build
export PKG_CONFIG_PATH="$PWD/build/_deps/ghostty-src/zig-out/share/pkgconfig"
```

Then, from this directory:

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
