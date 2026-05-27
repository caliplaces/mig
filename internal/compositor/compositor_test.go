package compositor_test

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/caliplaces/mig/internal/assets"
	"github.com/caliplaces/mig/internal/compositor"
)

func newLibrary(t *testing.T) *compositor.Library {
	t.Helper()
	lib, err := compositor.NewLibrary(assets.FS)
	if err != nil {
		t.Fatalf("NewLibrary: %v", err)
	}
	return lib
}

func TestLibrary_Groups_KnownNames(t *testing.T) {
	lib := newLibrary(t)
	groups := lib.Groups()
	if len(groups) == 0 {
		t.Fatal("expected non-empty muscle group list")
	}
	for _, want := range []string{"chest", "biceps", "triceps", "abs"} {
		if !lib.HasGroup(want) {
			t.Errorf("missing group %q", want)
		}
	}
}

func TestCompositor_Render_UnknownGroup(t *testing.T) {
	c := compositor.New(newLibrary(t))
	_, err := c.Render([]compositor.Highlight{{Group: "nonexistent"}}, false)
	if !errors.Is(err, compositor.ErrUnknownGroup) {
		t.Fatalf("expected ErrUnknownGroup, got %v", err)
	}
}

func TestCompositor_Render_EncodesToValidPNG(t *testing.T) {
	c := compositor.New(newLibrary(t))
	img, err := c.Render([]compositor.Highlight{
		{Group: "chest", Color: color.NRGBA{R: 0xff, A: 0xff}},
	}, false)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img.Bounds().Empty() {
		t.Fatal("rendered image has empty bounds")
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	if _, err := png.Decode(&buf); err != nil {
		t.Fatalf("png.Decode: %v", err)
	}
}

func TestCompositor_Render_RecolorAppliesToOutputPixels(t *testing.T) {
	c := compositor.New(newLibrary(t))
	red := color.NRGBA{R: 0xff, A: 0xff}
	img, err := c.Render([]compositor.Highlight{{Group: "chest", Color: red}}, true)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !hasPixelMatching(img, red) {
		t.Fatal("expected at least one red pixel in rendered image")
	}
}

func TestCompositor_Render_BaseIsNotMutated(t *testing.T) {
	c := compositor.New(newLibrary(t))
	beforeBase := c.Base(false)
	beforeSample := beforeBase.At(960, 960)

	_, err := c.Render([]compositor.Highlight{
		{Group: "chest", Color: color.NRGBA{R: 0xff, A: 0xff}},
	}, false)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	afterSample := c.Base(false).At(960, 960)
	if beforeSample != afterSample {
		t.Fatalf("base image mutated by Render: before=%v after=%v", beforeSample, afterSample)
	}
}

func hasPixelMatching(img image.Image, want color.NRGBA) bool {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if uint8(r>>8) == want.R && uint8(g>>8) == want.G &&
				uint8(bl>>8) == want.B && uint8(a>>8) == want.A {
				return true
			}
		}
	}
	return false
}
