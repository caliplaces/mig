package compositor

import (
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// ErrInvalidColor is returned by ParseColor for any input that cannot be parsed.
var ErrInvalidColor = errors.New("invalid color")

// ParseColor accepts:
//   - 6-char hex with optional '#' prefix: "FF0000", "#FF0000"
//   - 3-char shorthand hex: "F00", "#F00"
//   - comma-separated decimal RGB: "255,0,0"
//
// Whitespace is tolerated. Returns a fully opaque color.NRGBA.
func ParseColor(s string) (color.NRGBA, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return color.NRGBA{}, fmt.Errorf("%w: empty", ErrInvalidColor)
	}
	if strings.Contains(s, ",") {
		return parseRGB(s)
	}
	return parseHex(s)
}

func parseHex(s string) (color.NRGBA, error) {
	s = strings.TrimPrefix(s, "#")
	switch len(s) {
	case 3:
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	case 6:
		// ok
	default:
		return color.NRGBA{}, fmt.Errorf("%w: hex must be 3 or 6 chars, got %q", ErrInvalidColor, s)
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return color.NRGBA{}, fmt.Errorf("%w: %v", ErrInvalidColor, err)
	}
	return color.NRGBA{
		R: uint8(v >> 16),
		G: uint8(v >> 8),
		B: uint8(v),
		A: 0xff,
	}, nil
}

func parseRGB(s string) (color.NRGBA, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return color.NRGBA{}, fmt.Errorf("%w: rgb needs 3 components, got %d", ErrInvalidColor, len(parts))
	}
	out := color.NRGBA{A: 0xff}
	dst := []*uint8{&out.R, &out.G, &out.B}
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 || n > 255 {
			return color.NRGBA{}, fmt.Errorf("%w: rgb component %q out of range", ErrInvalidColor, p)
		}
		*dst[i] = uint8(n)
	}
	return out, nil
}
