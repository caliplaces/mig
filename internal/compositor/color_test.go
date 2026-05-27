package compositor

import (
	"errors"
	"image/color"
	"testing"
)

func TestParseColor(t *testing.T) {
	cases := []struct {
		in   string
		want color.NRGBA
		err  bool
	}{
		{"FF0000", color.NRGBA{R: 0xff, A: 0xff}, false},
		{"#FF0000", color.NRGBA{R: 0xff, A: 0xff}, false},
		{"#00FF00", color.NRGBA{G: 0xff, A: 0xff}, false},
		{"0000FF", color.NRGBA{B: 0xff, A: 0xff}, false},
		{"F00", color.NRGBA{R: 0xff, A: 0xff}, false},
		{"#abc", color.NRGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff}, false},
		{"255,0,0", color.NRGBA{R: 0xff, A: 0xff}, false},
		{"255, 0, 0", color.NRGBA{R: 0xff, A: 0xff}, false},
		{"0,0,0", color.NRGBA{A: 0xff}, false},
		{"  #ff0000  ", color.NRGBA{R: 0xff, A: 0xff}, false},

		{"", color.NRGBA{}, true},
		{"ZZZ", color.NRGBA{}, true},
		{"FF00", color.NRGBA{}, true},
		{"256,0,0", color.NRGBA{}, true},
		{"-1,0,0", color.NRGBA{}, true},
		{"1,2", color.NRGBA{}, true},
		{"a,b,c", color.NRGBA{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseColor(tc.in)
			if tc.err {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				if !errors.Is(err, ErrInvalidColor) {
					t.Fatalf("expected ErrInvalidColor, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
