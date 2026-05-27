// Package compositor renders anatomical images with arbitrary muscle
// groups highlighted in arbitrary colors. It is the only place that
// understands the image format conventions of the bundled PNG assets.
package compositor

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

// ErrUnknownGroup is returned by Render when a Highlight references a
// muscle group that does not exist in the Library.
var ErrUnknownGroup = errors.New("unknown muscle group")

// Highlight describes a single muscle group recolored with the given color.
type Highlight struct {
	Group string
	Color color.NRGBA
}

// Compositor renders muscle group overlays onto the base body image.
type Compositor struct {
	lib *Library
}

// New returns a Compositor backed by the given Library.
func New(lib *Library) *Compositor {
	return &Compositor{lib: lib}
}

// Groups exposes the available muscle group names.
func (c *Compositor) Groups() []string { return c.lib.Groups() }

// HasGroup reports whether the named group exists.
func (c *Compositor) HasGroup(name string) bool { return c.lib.HasGroup(name) }

// Base returns the (unaltered) base body image. The returned image must
// not be mutated by callers — it is shared across requests.
func (c *Compositor) Base(transparent bool) image.Image {
	return c.lib.baseFor(transparent)
}

// Render composes the given highlights onto the base image. Highlights
// are drawn in order; later entries overlay earlier ones.
func (c *Compositor) Render(highlights []Highlight, transparent bool) (image.Image, error) {
	src := c.lib.baseFor(transparent)
	out := image.NewNRGBA(src.Bounds())
	draw.Draw(out, out.Bounds(), src, src.Bounds().Min, draw.Src)

	for _, h := range highlights {
		m, ok := c.lib.muscleWithColor(h.Group, h.Color)
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrUnknownGroup, h.Group)
		}
		draw.Draw(out, m.Bounds(), m, m.Bounds().Min, draw.Over)
	}
	return out, nil
}
