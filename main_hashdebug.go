//go:build hashdebug

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

const maxLines = 24

// oldnums holds mapping from new line index to old line index, _NEWINDEX = -1.
var oldnums = make([]int, maxLines)

var (
	oldtext = make([]rune, maxLines)
	newtext = make([]rune, maxLines)
)

func resetState() {
	for i := 0; i < maxLines; i++ {
		oldnums[i] = -1
		oldtext[i] = '.'
		newtext[i] = '.'
	}
}

func dumpState(t *uv.Buffer) {
	oldStr := make([]rune, maxLines)
	newStr := make([]rune, maxLines)
	copy(oldStr, oldtext)
	copy(newStr, newtext)

	// clear buffer
	t.Clear()

	for i, r := range string(oldStr) {
		var c uv.Cell
		c.Content = string(r)
		c.Width = 1
		t.SetCell(i, 0, &c)
	}

	for i, r := range string(newStr) {
		var c uv.Cell
		c.Content = string(r)
		c.Width = 1
		t.SetCell(i, 2, &c)
	}

	fmt.Println("Old lines: [" + string(oldStr) + "]")
	fmt.Println("New lines: [" + string(newStr) + "]")
}

func usage() {
	msg := []string{
		"hashmap test-driver (Go / Ultraviolet)",
		"",
		"#  comment",
		"l  get initial line number vector",
		"n  use following letters as text of new lines",
		"o  use following letters as text of old lines",
		"d  dump state of test arrays (Ultraviolet renderer)",
		"?  this message",
	}
	for _, s := range msg {
		fmt.Fprintln(os.Stderr, s)
	}
}

func init() {
	flag.Usage = usage
}

func main() {
	flag.Parse()

	t := uv.NewBuffer(80, maxLines)

	resetState()

	if fi, _ := os.Stdin.Stat(); (fi.Mode() & os.ModeCharDevice) != 0 {
		usage()
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		switch line[0] {
		case '#':
			fmt.Fprintln(os.Stderr, line)
		case 'l':
			for i := 0; i < maxLines; i++ {
				oldnums[i] = -1
			}
			fields := strings.Fields(line[1:])
			for i, f := range fields {
				v, err := strconv.Atoi(f)
				if err == nil && i < maxLines {
					oldnums[i] = v
				}
			}
		case 'n':
			for i := 0; i < maxLines; i++ {
				newtext[i] = '.'
			}
			runes := []rune(line[1:])
			for i := 0; i < len(runes) && i < maxLines; i++ {
				if runes[i] == '\n' {
					break
				}
				newtext[i] = runes[i]
			}
		case 'o':
			for i := 0; i < maxLines; i++ {
				oldtext[i] = '.'
			}
			runes := []rune(line[1:])
			for i := 0; i < len(runes) && i < maxLines; i++ {
				if runes[i] == '\n' {
					break
				}
				oldtext[i] = runes[i]
			}
		case 'd':
			dumpState(t)
		case '?':
			usage()
		default:
			usage()
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}
