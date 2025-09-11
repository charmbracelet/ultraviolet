package uv

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/charmbracelet/colorprofile"
)

func TestConvertStyle(t *testing.T) {
	s := Style{
		Fg: color.Black,
		Bg: color.White,
		Ul: color.Black,
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
				Fg: colorprofile.ANSI256.Convert(color.Black),
				Bg: colorprofile.ANSI256.Convert(color.White),
				Ul: colorprofile.ANSI256.Convert(color.Black),
			},
		},
		{
			name:    "16 Color",
			profile: colorprofile.ANSI,
			want: Style{
				Fg: colorprofile.ANSI.Convert(color.Black),
				Bg: colorprofile.ANSI.Convert(color.White),
				Ul: colorprofile.ANSI.Convert(color.Black),
			},
		},
		{
			name:    "Grayscale",
			profile: colorprofile.Ascii,
			want: Style{
				Fg: nil,
				Bg: nil,
				Ul: nil,
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
