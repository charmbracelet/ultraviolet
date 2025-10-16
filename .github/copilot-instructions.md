# Ultraviolet Terminal UI Toolkit

Ultraviolet is a Go library for creating terminal-based user interfaces (TUIs). It provides primitives for manipulating terminal emulators with a focus on cell-based rendering, input handling, and cross-platform compatibility.

**ALWAYS reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.**

## Working Effectively

### Prerequisites and Setup
- **Go 1.24.0 or later required** (project uses Go 1.24.0 minimum)
- Install golangci-lint latest version (v2.4.0+) for Go 1.24 support:
  ```bash
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
  export PATH=$PATH:$(go env GOPATH)/bin
  ```

### Core Development Workflow
Run these commands in sequence for a complete development setup:

1. **Download dependencies:**
   ```bash
   go mod download && go mod tidy
   ```
   - Takes ~2-3 seconds on first run

2. **Build the library:**
   ```bash
   go build ./...
   ```
   - Takes ~8 seconds. NEVER CANCEL. Set timeout to 30+ seconds.

3. **Build examples:**
   ```bash
   cd examples && go build ./...
   ```
   - Takes ~3 seconds. NEVER CANCEL. Set timeout to 15+ seconds.

4. **Run tests:**
   ```bash
   go test ./...
   ```
   - Takes ~3-4 seconds. NEVER CANCEL. Set timeout to 30+ seconds.

5. **Run tests with race detection:**
   ```bash
   go test -race ./...
   ```
   - Takes ~14 seconds. NEVER CANCEL. Set timeout to 60+ seconds.

6. **Run tests with coverage:**
   ```bash
   go test -cover ./...
   ```
   - Takes ~3 seconds. Current coverage: ~56.6%

7. **Run linting:**
   ```bash
   golangci-lint run
   ```
   - Takes ~3 seconds. NEVER CANCEL. Set timeout to 30+ seconds.

8. **Run benchmarks:**
   ```bash
   go test -bench=. ./...
   ```
   - Takes ~3 seconds. NEVER CANCEL. Set timeout to 30+ seconds.

### Validation Requirements
**CRITICAL**: Always run these validation steps after making changes:

1. **Build validation:**
   ```bash
   go build ./... && cd examples && go build ./...
   ```

2. **Test validation:**
   ```bash
   go test ./... && go test -race ./...
   ```

3. **Lint validation:**
   ```bash
   golangci-lint run
   ```

4. **Example validation:**
   Build and test at least one example to ensure functionality:
   ```bash
   cd examples/helloworld
   go build -o helloworld
   timeout 2s ./helloworld || true  # Should display "Hello, World!"
   ```

### Advanced Testing

**Fuzz testing available:**
```bash
go test -fuzz=FuzzParseSequence -fuzztime=30s
```
- Takes ~30+ seconds for meaningful fuzzing. NEVER CANCEL. Set timeout to 120+ seconds.

## Repository Structure

### Key Directories
- **Root directory**: Contains 45 Go source files with core functionality
- **examples/**: 11 example applications demonstrating usage patterns:
  - `helloworld/`: Basic terminal UI example
  - `altscreen/`: Alternate screen buffer usage
  - `draw/`: Drawing and rendering examples
  - `image/`: Image rendering in terminal
  - `layout/`: Layout management
  - `tv/`: Complex TUI example
  - `panic/`, `prependline/`, `space/`: Specific feature demos
- **screen/**: Screen abstraction package
- **.github/workflows/**: CI/CD pipelines

### Critical Files
- **`uv.go`**: Main package entry point and core interfaces
- **`terminal.go`**: Terminal manipulation and control
- **`terminal_renderer.go`**: Core rendering engine ("The Cursed Renderer")
- **`buffer.go`**: Cell-based buffer management  
- **`decoder.go`**: Input event parsing and decoding
- **`key.go`** / **`key_table.go`**: Keyboard input handling
- **`event.go`**: Event system definitions
- **`cell.go`**: Terminal cell abstraction
- **`styled.go`**: Styled text rendering
- **Platform-specific files**: `*_windows.go`, `*_unix.go`, `*_other.go`

### Build Constraints and Platform Support
The codebase uses extensive platform-specific code with build constraints:
- **Windows-specific**: `*_windows.go` files
- **Unix-specific**: `*_unix.go` files  
- **BSD-specific**: `*_bsdly.go` files
- **Cross-platform fallbacks**: `*_other.go` files

**Platform limitations**: Some features like window size notifications are not supported on all platforms (see `winch_other.go`).

## Development Patterns

### Testing Patterns
- **Unit tests**: Standard `*_test.go` files with good coverage
- **Benchmark tests**: Performance testing for critical paths
- **Output tests**: `terminal_renderer_output_test.go` for rendering validation
- **Race detection**: Always test with `-race` flag for concurrent code
- **Fuzz testing**: `FuzzParseSequence` for input parsing robustness

### Code Organization
- **Event-driven architecture**: Central event system in `event.go`
- **Cell-based rendering**: Everything operates on terminal cells
- **Platform abstraction**: Clean separation of platform-specific code
- **Interface-driven design**: Heavy use of interfaces for extensibility

### Important Gotchas
- **API instability warning**: Per README, expect no stability guarantees currently
- **Go 1.24+ requirement**: Newer golangci-lint versions needed
- **Platform differences**: Window management varies significantly by OS
- **Terminal compatibility**: Complex terminal capability detection
- **Input parsing complexity**: See notes about F3 key sequences and cursor position reports

## Common Development Tasks

### Adding New Features
1. **Identify the right file**: Use the structure above to find where new code belongs
2. **Check platform requirements**: Determine if platform-specific code is needed
3. **Add tests**: Follow existing test patterns in `*_test.go` files
4. **Update examples**: Add or modify examples to demonstrate new features
5. **Run full validation**: Complete build, test, and lint cycle

### Debugging Terminal Issues
1. **Enable debug logging**: Examples create `uv_debug.log` files
2. **Test with different terminals**: Check compatibility across terminal types
3. **Use output tests**: Add rendering tests in `terminal_renderer_output_test.go`
4. **Check platform-specific code**: Review appropriate `*_windows.go` or `*_unix.go` files

### Working with Examples
Examples are the best way to understand and test functionality:
```bash
cd examples/[example_name]
go build -o [example_name]
./[example_name]
```

**Interactive testing limitation**: Examples require real terminal interaction - automated testing is limited.

### Performance Optimization
- **Profile with benchmarks**: Use `go test -bench=. -cpuprofile=cpu.prof`
- **Focus on rendering**: The cell-based diff algorithm is performance-critical
- **Memory allocation**: Pay attention to allocations in hot paths
- **Terminal I/O**: Minimize escape sequence generation

## CI/CD Integration

The project uses GitHub Actions workflows:
- **build.yml**: Multi-version Go builds including examples
- **lint.yml**: Linting with golangci-lint
- **coverage.yml**: Test coverage reporting
- **examples.yml**: Example application testing

**Always ensure your changes pass all CI checks before submitting.**

## Testing Scenarios

**After making changes, always test these scenarios:**

1. **Basic functionality validation:**
   ```bash
   cd examples/helloworld && go build -o helloworld && timeout 2s ./helloworld
   # Should output: "Hello, World!"
   ```

2. **Interactive mode validation:**
   ```bash
   cd examples/altscreen && go build -o altscreen && timeout 1s ./altscreen
   # Should output inline mode message
   ```

3. **Complex rendering validation:**
   ```bash
   cd examples/draw && go build -o draw && timeout 1s ./draw
   # Should show drawing interface messages
   ```

## Quick Reference Commands

```bash
# Complete development cycle (recommended for all changes)
export PATH=$PATH:$(go env GOPATH)/bin
go mod tidy && go build ./... && cd examples && go build ./... && cd .. && go test ./... && golangci-lint run

# Fast iteration cycle  
go build ./... && go test ./...

# Example testing pattern
cd examples/[name] && go build -o [name] && timeout 2s ./[name]

# Coverage with output
go test -cover -coverprofile=coverage.out ./...

# Race detection (slower but important)
go test -race ./...

# Fuzzing (run periodically)
go test -fuzz=FuzzParseSequence -fuzztime=10s

# Documentation check
go doc .
```

**REMINDER**: NEVER CANCEL long-running commands. Builds may take up to 30 seconds, race detection up to 60 seconds, and fuzzing should run for meaningful duration.