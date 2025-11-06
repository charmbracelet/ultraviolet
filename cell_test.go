package uv

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/charmbracelet/colorprofile"
)

func TestConvertStyle(t *testing.T) {
	s := Style{
		Fg:             color.Black,
		Bg:             color.White,
		UnderlineColor: color.Black,
	}
	tests := []struct {
		name    string
		profile colorprofile.Profile
		want    Style
	}{
		{
			name:    "True Color",
			profile: colorprofile.TrueColor,
			want:    s,
		},
		{
			name:    "256 Color",
			profile: colorprofile.ANSI256,
			want: Style{
				Fg:             colorprofile.ANSI256.Convert(color.Black),
				Bg:             colorprofile.ANSI256.Convert(color.White),
				UnderlineColor: colorprofile.ANSI256.Convert(color.Black),
			},
		},
		{
			name:    "16 Color",
			profile: colorprofile.ANSI,
			want: Style{
				Fg:             colorprofile.ANSI.Convert(color.Black),
				Bg:             colorprofile.ANSI.Convert(color.White),
				UnderlineColor: colorprofile.ANSI.Convert(color.Black),
			},
		},
		{
			name:    "Grayscale",
			profile: colorprofile.Ascii,
			want: Style{
				Fg:             nil,
				Bg:             nil,
				UnderlineColor: nil,
			},
		},
		{
			name:    "No Profile",
			profile: colorprofile.NoTTY,
			want:    Style{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertStyle(s, tt.profile); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertStyle() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConvertLink(t *testing.T) {
	l := Link{
		URL:    "https://example.com",
		Params: "id=1",
	}
	tests := []struct {
		name    string
		profile colorprofile.Profile
		want    Link
	}{
		{
			name:    "True Color",
			profile: colorprofile.TrueColor,
			want:    l,
		},
		{
			name:    "No TTY",
			profile: colorprofile.NoTTY,
			want:    Link{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertLink(l, tt.profile); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertLink() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestStyleDiff(t *testing.T) {
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	green := color.RGBA{0, 255, 0, 255}
	yellow := color.RGBA{255, 255, 0, 255}
	cyan := color.RGBA{0, 255, 255, 255}
	magenta := color.RGBA{255, 0, 255, 255}

	tests := []struct {
		name string
		from *Style
		to   *Style
		want string
	}{
		// Nil and zero cases
		{
			name: "both nil",
			from: nil,
			to:   nil,
			want: "",
		},
		{
			name: "from nil to zero",
			from: nil,
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "from zero to zero",
			from: &Style{},
			to:   &Style{},
			want: "",
		},
		{
			name: "from nil to styled",
			from: nil,
			to:   &Style{Fg: red, Attrs: AttrBold},
			want: "\x1b[1;38;2;255;0;0m",
		},

		// Foreground color tests
		{
			name: "foreground color change",
			from: &Style{Fg: red},
			to:   &Style{Fg: blue},
			want: "\x1b[38;2;0;0;255m",
		},
		{
			name: "add foreground color",
			from: &Style{},
			to:   &Style{Fg: red},
			want: "\x1b[38;2;255;0;0m",
		},
		{
			name: "remove foreground color",
			from: &Style{Fg: red},
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "foreground color same",
			from: &Style{Fg: red},
			to:   &Style{Fg: red},
			want: "",
		},

		// Background color tests
		{
			name: "background color change",
			from: &Style{Bg: red},
			to:   &Style{Bg: blue},
			want: "\x1b[48;2;0;0;255m",
		},
		{
			name: "add background color",
			from: &Style{},
			to:   &Style{Bg: blue},
			want: "\x1b[48;2;0;0;255m",
		},
		{
			name: "remove background color",
			from: &Style{Bg: blue},
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "background color same",
			from: &Style{Bg: blue},
			to:   &Style{Bg: blue},
			want: "",
		},

		// Underline color tests
		{
			name: "underline color change",
			from: &Style{UnderlineColor: red, Underline: UnderlineStyleSingle},
			to:   &Style{UnderlineColor: blue, Underline: UnderlineStyleSingle},
			want: "\x1b[58;2;0;0;255m",
		},
		{
			name: "add underline color",
			from: &Style{Underline: UnderlineStyleSingle},
			to:   &Style{UnderlineColor: green, Underline: UnderlineStyleSingle},
			want: "\x1b[58;2;0;255;0m",
		},
		{
			name: "remove underline color",
			from: &Style{UnderlineColor: green, Underline: UnderlineStyleSingle},
			to:   &Style{Underline: UnderlineStyleSingle},
			want: "\x1b[59m",
		},
		{
			name: "underline color same",
			from: &Style{UnderlineColor: green, Underline: UnderlineStyleSingle},
			to:   &Style{UnderlineColor: green, Underline: UnderlineStyleSingle},
			want: "",
		},

		// Bold attribute tests
		{
			name: "add bold",
			from: &Style{},
			to:   &Style{Attrs: AttrBold},
			want: "\x1b[1m",
		},
		{
			name: "remove bold",
			from: &Style{Attrs: AttrBold},
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "keep bold",
			from: &Style{Attrs: AttrBold},
			to:   &Style{Attrs: AttrBold},
			want: "",
		},

		// Faint attribute tests
		{
			name: "add faint",
			from: &Style{},
			to:   &Style{Attrs: AttrFaint},
			want: "\x1b[2m",
		},
		{
			name: "remove faint",
			from: &Style{Attrs: AttrFaint},
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "keep faint",
			from: &Style{Attrs: AttrFaint},
			to:   &Style{Attrs: AttrFaint},
			want: "",
		},
		{
			name: "bold to faint",
			from: &Style{Attrs: AttrBold},
			to:   &Style{Attrs: AttrFaint},
			want: "\x1b[22;2m",
		},
		{
			name: "faint to bold",
			from: &Style{Attrs: AttrFaint},
			to:   &Style{Attrs: AttrBold},
			want: "\x1b[22;1m",
		},
		{
			name: "bold and faint to bold",
			from: &Style{Attrs: AttrBold | AttrFaint},
			to:   &Style{Attrs: AttrBold},
			want: "\x1b[22;1m",
		},
		{
			name: "bold to bold and faint",
			from: &Style{Attrs: AttrBold},
			to:   &Style{Attrs: AttrBold | AttrFaint},
			want: "\x1b[2m",
		},

		// Italic attribute tests
		{
			name: "add italic",
			from: &Style{},
			to:   &Style{Attrs: AttrItalic},
			want: "\x1b[3m",
		},
		{
			name: "remove italic",
			from: &Style{Attrs: AttrItalic},
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "keep italic",
			from: &Style{Attrs: AttrItalic},
			to:   &Style{Attrs: AttrItalic},
			want: "",
		},

		// Bold and Italic combination tests
		{
			name: "bold to bold and italic",
			from: &Style{Attrs: AttrBold},
			to:   &Style{Attrs: AttrBold | AttrItalic},
			want: "\x1b[3m",
		},
		{
			name: "bold and italic to bold",
			from: &Style{Attrs: AttrBold | AttrItalic},
			to:   &Style{Attrs: AttrBold},
			want: "\x1b[23m",
		},

		// Bold, Faint, and Italic combination tests
		{
			name: "bold and faint to italic",
			from: &Style{Attrs: AttrBold | AttrFaint},
			to:   &Style{Attrs: AttrItalic},
			want: "\x1b[22;3m",
		},
		{
			name: "italic to bold and faint",
			from: &Style{Attrs: AttrItalic},
			to:   &Style{Attrs: AttrBold | AttrFaint},
			want: "\x1b[23;1;2m",
		},
		{
			name: "bold, faint, and italic to bold",
			from: &Style{Attrs: AttrBold | AttrFaint | AttrItalic},
			to:   &Style{Attrs: AttrBold},
			want: "\x1b[22;23;1m",
		},
		{
			name: "bold to bold, faint, and italic",
			from: &Style{Attrs: AttrBold},
			to:   &Style{Attrs: AttrBold | AttrFaint | AttrItalic},
			want: "\x1b[2;3m",
		},
		{
			name: "faint to bold and italic",
			from: &Style{Attrs: AttrFaint},
			to:   &Style{Attrs: AttrBold | AttrItalic},
			want: "\x1b[22;1;3m",
		},
		{
			name: "italic to bold and faint",
			from: &Style{Attrs: AttrItalic},
			to:   &Style{Attrs: AttrBold | AttrFaint},
			want: "\x1b[23;1;2m",
		},
		{
			name: "bold, faint, and italic to faint",
			from: &Style{Attrs: AttrBold | AttrFaint | AttrItalic},
			to:   &Style{Attrs: AttrFaint},
			want: "\x1b[22;23;2m",
		},
		{
			name: "bold, faint, and italic to italic",
			from: &Style{Attrs: AttrBold | AttrFaint | AttrItalic},
			to:   &Style{Attrs: AttrItalic},
			want: "\x1b[22m",
		},
		{
			name: "faint to bold, faint, and italic",
			from: &Style{Attrs: AttrFaint},
			to:   &Style{Attrs: AttrBold | AttrFaint | AttrItalic},
			want: "\x1b[1;3m",
		},
		{
			name: "italic to bold, faint, and italic",
			from: &Style{Attrs: AttrItalic},
			to:   &Style{Attrs: AttrBold | AttrFaint | AttrItalic},
			want: "\x1b[1;2m",
		},
		{
			name: "bold, faint, and italic to bold, faint, and italic",
			from: &Style{Attrs: AttrBold | AttrFaint | AttrItalic},
			to:   &Style{Attrs: AttrBold | AttrFaint | AttrItalic},
			want: "",
		},
		{
			name: "no attributes to bold, faint, and italic",
			from: &Style{},
			to:   &Style{Attrs: AttrBold | AttrFaint | AttrItalic},
			want: "\x1b[1;2;3m",
		},

		// Slow blink attribute tests
		{
			name: "add slow blink",
			from: &Style{},
			to:   &Style{Attrs: AttrBlink},
			want: "\x1b[5m",
		},
		{
			name: "remove slow blink",
			from: &Style{Attrs: AttrBlink},
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "keep slow blink",
			from: &Style{Attrs: AttrBlink},
			to:   &Style{Attrs: AttrBlink},
			want: "",
		},

		// Rapid blink attribute tests
		{
			name: "add rapid blink",
			from: &Style{},
			to:   &Style{Attrs: AttrRapidBlink},
			want: "\x1b[6m",
		},
		{
			name: "remove rapid blink",
			from: &Style{Attrs: AttrRapidBlink},
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "keep rapid blink",
			from: &Style{Attrs: AttrRapidBlink},
			to:   &Style{Attrs: AttrRapidBlink},
			want: "",
		},
		{
			name: "change from slow to rapid blink",
			from: &Style{Attrs: AttrBlink},
			to:   &Style{Attrs: AttrRapidBlink},
			want: "\x1b[25;6m",
		},
		{
			name: "change from rapid to slow blink",
			from: &Style{Attrs: AttrRapidBlink},
			to:   &Style{Attrs: AttrBlink},
			want: "\x1b[25;5m",
		},
		{
			name: "slow and rapid blink to slow blink",
			from: &Style{Attrs: AttrBlink | AttrRapidBlink},
			to:   &Style{Attrs: AttrBlink},
			want: "\x1b[25;5m",
		},

		// Reverse attribute tests
		{
			name: "add reverse",
			from: &Style{},
			to:   &Style{Attrs: AttrReverse},
			want: "\x1b[7m",
		},
		{
			name: "remove reverse",
			from: &Style{Attrs: AttrReverse},
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "keep reverse",
			from: &Style{Attrs: AttrReverse},
			to:   &Style{Attrs: AttrReverse},
			want: "",
		},

		// Conceal attribute tests
		{
			name: "add conceal",
			from: &Style{},
			to:   &Style{Attrs: AttrConceal},
			want: "\x1b[8m",
		},
		{
			name: "remove conceal",
			from: &Style{Attrs: AttrConceal},
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "keep conceal",
			from: &Style{Attrs: AttrConceal},
			to:   &Style{Attrs: AttrConceal},
			want: "",
		},

		// Strikethrough attribute tests
		{
			name: "add strikethrough",
			from: &Style{},
			to:   &Style{Attrs: AttrStrikethrough},
			want: "\x1b[9m",
		},
		{
			name: "remove strikethrough",
			from: &Style{Attrs: AttrStrikethrough},
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "keep strikethrough",
			from: &Style{Attrs: AttrStrikethrough},
			to:   &Style{Attrs: AttrStrikethrough},
			want: "",
		},

		// Underline style tests
		{
			name: "add single underline",
			from: &Style{},
			to:   &Style{Underline: UnderlineStyleSingle},
			want: "\x1b[4m",
		},
		{
			name: "add double underline",
			from: &Style{},
			to:   &Style{Underline: UnderlineStyleDouble},
			want: "\x1b[4:2m",
		},
		{
			name: "add curly underline",
			from: &Style{},
			to:   &Style{Underline: UnderlineStyleCurly},
			want: "\x1b[4:3m",
		},
		{
			name: "add dotted underline",
			from: &Style{},
			to:   &Style{Underline: UnderlineStyleDotted},
			want: "\x1b[4:4m",
		},
		{
			name: "add dashed underline",
			from: &Style{},
			to:   &Style{Underline: UnderlineStyleDashed},
			want: "\x1b[4:5m",
		},
		{
			name: "change underline style single to double",
			from: &Style{Underline: UnderlineStyleSingle},
			to:   &Style{Underline: UnderlineStyleDouble},
			want: "\x1b[4:2m",
		},
		{
			name: "change underline style double to curly",
			from: &Style{Underline: UnderlineStyleDouble},
			to:   &Style{Underline: UnderlineStyleCurly},
			want: "\x1b[4:3m",
		},
		{
			name: "remove underline",
			from: &Style{Underline: UnderlineStyleSingle},
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "keep underline style",
			from: &Style{Underline: UnderlineStyleSingle},
			to:   &Style{Underline: UnderlineStyleSingle},
			want: "",
		},

		// Multiple attribute combinations
		{
			name: "add multiple attributes",
			from: &Style{},
			to:   &Style{Attrs: AttrBold | AttrItalic, Underline: UnderlineStyleSingle},
			want: "\x1b[1;3;4m",
		},
		{
			name: "remove multiple attributes",
			from: &Style{Attrs: AttrBold | AttrItalic | AttrReverse},
			to:   &Style{},
			want: "\x1b[m",
		},
		{
			name: "combine multiple attribute changes",
			from: &Style{Attrs: AttrBold | AttrItalic},
			to:   &Style{Attrs: AttrBold | AttrReverse},
			want: "\x1b[23;7m",
		},
		{
			name: "swap italic and strikethrough",
			from: &Style{Attrs: AttrItalic},
			to:   &Style{Attrs: AttrStrikethrough},
			want: "\x1b[23;9m",
		},
		{
			name: "all attributes added",
			from: &Style{},
			to:   &Style{Attrs: AttrBold | AttrFaint | AttrItalic | AttrBlink | AttrRapidBlink | AttrReverse | AttrConceal | AttrStrikethrough},
			want: "\x1b[1;2;3;5;6;7;8;9m",
		},
		{
			name: "all attributes removed",
			from: &Style{Attrs: AttrBold | AttrFaint | AttrItalic | AttrBlink | AttrRapidBlink | AttrReverse | AttrConceal | AttrStrikethrough},
			to:   &Style{},
			want: "\x1b[m",
		},

		// Complex style changes with colors and attributes
		{
			name: "complex style change with all properties",
			from: &Style{
				Fg:    red,
				Bg:    blue,
				Attrs: AttrBold,
			},
			to: &Style{
				Fg:             green,
				Bg:             yellow,
				UnderlineColor: cyan,
				Attrs:          AttrItalic,
				Underline:      UnderlineStyleSingle,
			},
			want: "\x1b[38;2;0;255;0;48;2;255;255;0;58;2;0;255;255;22;3;4m",
		},
		{
			name: "complex change keeping some properties",
			from: &Style{
				Fg:        red,
				Bg:        blue,
				Attrs:     AttrBold | AttrItalic,
				Underline: UnderlineStyleSingle,
			},
			to: &Style{
				Fg:        red,
				Bg:        green,
				Attrs:     AttrBold | AttrReverse,
				Underline: UnderlineStyleDouble,
			},
			want: "\x1b[48;2;0;255;0;23;7;4:2m",
		},
		{
			name: "complete style reset",
			from: &Style{
				Fg:             red,
				Bg:             blue,
				UnderlineColor: green,
				Attrs:          AttrBold | AttrItalic | AttrReverse,
				Underline:      UnderlineStyleSingle,
			},
			to:   &Style{},
			want: "\x1b[m",
		},

		// Edge cases
		{
			name: "no changes with all properties",
			from: &Style{
				Fg:             red,
				Bg:             blue,
				UnderlineColor: green,
				Attrs:          AttrBold | AttrItalic,
				Underline:      UnderlineStyleSingle,
			},
			to: &Style{
				Fg:             red,
				Bg:             blue,
				UnderlineColor: green,
				Attrs:          AttrBold | AttrItalic,
				Underline:      UnderlineStyleSingle,
			},
			want: "",
		},
		{
			name: "only colors change",
			from: &Style{
				Fg:    red,
				Bg:    blue,
				Attrs: AttrBold,
			},
			to: &Style{
				Fg:    green,
				Bg:    yellow,
				Attrs: AttrBold,
			},
			want: "\x1b[38;2;0;255;0;48;2;255;255;0m",
		},
		{
			name: "only attributes change",
			from: &Style{
				Fg:    red,
				Attrs: AttrBold,
			},
			to: &Style{
				Fg:    red,
				Attrs: AttrItalic,
			},
			want: "\x1b[22;3m",
		},
		{
			name: "add all colors",
			from: &Style{},
			to: &Style{
				Fg:             red,
				Bg:             blue,
				UnderlineColor: green,
				Underline:      UnderlineStyleSingle,
			},
			want: "\x1b[38;2;255;0;0;48;2;0;0;255;58;2;0;255;0;4m",
		},
		{
			name: "add all colors without underline",
			from: &Style{},
			to: &Style{
				Fg:             red,
				Bg:             blue,
				UnderlineColor: green,
			},
			want: "\x1b[38;2;255;0;0;48;2;0;0;255;58;2;0;255;0m",
		},
		{
			name: "remove all colors with attributes",
			from: &Style{
				Fg:    red,
				Bg:    blue,
				Attrs: AttrBold,
			},
			to: &Style{
				Attrs: AttrBold,
			},
			want: "\x1b[39;49m",
		},
		{
			name: "change all colors",
			from: &Style{
				Fg:             red,
				Bg:             blue,
				UnderlineColor: green,
			},
			to: &Style{
				Fg:             cyan,
				Bg:             magenta,
				UnderlineColor: yellow,
			},
			want: "\x1b[38;2;0;255;255;48;2;255;0;255;58;2;255;255;0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StyleDiff(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("StyleDiff() = %q, want %q", got, tt.want)
			}
		})
	}
}
