package compositor

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"sort"
	"strings"
)

// HighlightColor is the RGB value used in the source muscle PNGs to mark
// the colorable region. At render time the palette entry matching this
// color is swapped for the requested color.
var HighlightColor = color.NRGBA{R: 89, G: 136, B: 255, A: 0xff}

// Library holds preloaded PNG assets and pre-computed palette indices,
// so rendering a request never touches disk and recoloring is O(1).
type Library struct {
	base            image.Image
	baseTransparent image.Image
	muscles         map[string]*muscleAsset
	groups          []string
}

type muscleAsset struct {
	img            *image.Paletted
	highlightIndex int // -1 if the highlight color is not present
}

// NewLibrary loads every PNG under "images/" in the given fs.FS. It
// expects baseImage.png and baseImage_transparent.png plus one PNG per
// muscle group. Files that are not paletted (palette-indexed) are rejected.
func NewLibrary(filesystem fs.FS) (*Library, error) {
	base, err := decodePNG(filesystem, "images/baseImage.png")
	if err != nil {
		return nil, fmt.Errorf("load baseImage: %w", err)
	}
	baseT, err := decodePNG(filesystem, "images/baseImage_transparent.png")
	if err != nil {
		return nil, fmt.Errorf("load baseImage_transparent: %w", err)
	}

	entries, err := fs.ReadDir(filesystem, "images")
	if err != nil {
		return nil, fmt.Errorf("read images dir: %w", err)
	}

	lib := &Library{
		base:            base,
		baseTransparent: baseT,
		muscles:         make(map[string]*muscleAsset),
	}

	skip := map[string]struct{}{
		"baseImage.png":             {},
		"baseImage_transparent.png": {},
		"background.png":            {},
	}

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".png") {
			continue
		}
		if _, ok := skip[name]; ok {
			continue
		}
		img, err := decodePNG(filesystem, "images/"+name)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", name, err)
		}
		pal, ok := img.(*image.Paletted)
		if !ok {
			return nil, fmt.Errorf("%s: expected paletted PNG, got %T", name, img)
		}
		group := strings.TrimSuffix(name, ".png")
		lib.muscles[group] = &muscleAsset{
			img:            pal,
			highlightIndex: findPaletteIndex(pal.Palette, HighlightColor),
		}
		lib.groups = append(lib.groups, group)
	}
	sort.Strings(lib.groups)
	return lib, nil
}

func decodePNG(filesystem fs.FS, path string) (image.Image, error) {
	f, err := filesystem.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func findPaletteIndex(p color.Palette, target color.NRGBA) int {
	for i, c := range p {
		r, g, b, a := c.RGBA()
		// color.Color.RGBA returns 16-bit values; shift back to 8-bit.
		if uint8(r>>8) == target.R && uint8(g>>8) == target.G &&
			uint8(b>>8) == target.B && uint8(a>>8) == target.A {
			return i
		}
	}
	return -1
}

// Groups returns the sorted list of available muscle groups.
func (l *Library) Groups() []string {
	out := make([]string, len(l.groups))
	copy(out, l.groups)
	return out
}

// HasGroup reports whether the named group exists.
func (l *Library) HasGroup(name string) bool {
	_, ok := l.muscles[name]
	return ok
}

// muscleWithColor returns a paletted image sharing the original pixel
// buffer but with a freshly cloned palette where the highlight entry has
// been swapped for c. Safe for concurrent reads because Pix is never mutated.
func (l *Library) muscleWithColor(group string, c color.NRGBA) (*image.Paletted, bool) {
	m, ok := l.muscles[group]
	if !ok {
		return nil, false
	}
	pal := make(color.Palette, len(m.img.Palette))
	copy(pal, m.img.Palette)
	if m.highlightIndex >= 0 {
		pal[m.highlightIndex] = c
	}
	return &image.Paletted{
		Pix:     m.img.Pix,
		Stride:  m.img.Stride,
		Rect:    m.img.Rect,
		Palette: pal,
	}, true
}

func (l *Library) baseFor(transparent bool) image.Image {
	if transparent {
		return l.baseTransparent
	}
	return l.base
}
