package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"net/http"
	"strconv"
	"strings"

	"github.com/caliplaces/mig/internal/compositor"
	"github.com/caliplaces/mig/internal/docs"
)

type handlers struct {
	c *compositor.Compositor
}

func (h *handlers) root(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("muscle group image generator — see /getMuscleGroups\n"))
}

func (h *handlers) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *handlers) openapiSpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = w.Write(docs.Spec)
}

func (h *handlers) swaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(docs.SwaggerUI))
}

func (h *handlers) getMuscleGroups(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.c.Groups())
}

func (h *handlers) getBaseImage(w http.ResponseWriter, r *http.Request) {
	transparent := boolParam(r, "transparentBackground", false)
	writePNG(w, h.c.Base(transparent))
}

func (h *handlers) getImage(w http.ResponseWriter, r *http.Request) {
	groupsParam := strings.TrimSpace(r.URL.Query().Get("muscleGroups"))
	if groupsParam == "" {
		writeError(w, http.StatusBadRequest, "muscleGroups is required")
		return
	}
	transparent := boolParam(r, "transparentBackground", false)
	colorParam := r.URL.Query().Get("color")

	col := compositor.HighlightColor
	if colorParam != "" {
		c, err := compositor.ParseColor(colorParam)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		col = c
	}

	groups := splitCSV(groupsParam)
	if err := validateGroups(h.c, groups); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	highlights := make([]compositor.Highlight, len(groups))
	for i, g := range groups {
		highlights[i] = compositor.Highlight{Group: g, Color: col}
	}
	renderAndWrite(w, h.c, highlights, transparent)
}

func (h *handlers) getMulticolorImage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	pGroups := splitCSV(q.Get("primaryMuscleGroups"))
	sGroups := splitCSV(q.Get("secondaryMuscleGroups"))
	pColor := q.Get("primaryColor")
	sColor := q.Get("secondaryColor")

	if len(pGroups) == 0 || len(sGroups) == 0 || pColor == "" || sColor == "" {
		writeError(w, http.StatusBadRequest,
			"primaryMuscleGroups, secondaryMuscleGroups, primaryColor, secondaryColor are required")
		return
	}
	transparent := boolParam(r, "transparentBackground", false)

	pCol, err := compositor.ParseColor(pColor)
	if err != nil {
		writeError(w, http.StatusBadRequest, "primaryColor: "+err.Error())
		return
	}
	sCol, err := compositor.ParseColor(sColor)
	if err != nil {
		writeError(w, http.StatusBadRequest, "secondaryColor: "+err.Error())
		return
	}

	all := make([]string, 0, len(pGroups)+len(sGroups))
	all = append(all, pGroups...)
	all = append(all, sGroups...)
	if err := validateGroups(h.c, all); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	highlights := make([]compositor.Highlight, 0, len(pGroups)+len(sGroups))
	for _, g := range pGroups {
		highlights = append(highlights, compositor.Highlight{Group: g, Color: pCol})
	}
	for _, g := range sGroups {
		highlights = append(highlights, compositor.Highlight{Group: g, Color: sCol})
	}
	renderAndWrite(w, h.c, highlights, transparent)
}

func (h *handlers) getIndividualColorImage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	groups := splitCSV(q.Get("muscleGroups"))
	colors := splitCSV(q.Get("colors"))

	if len(groups) == 0 || len(colors) == 0 {
		writeError(w, http.StatusBadRequest, "muscleGroups and colors are required")
		return
	}
	if len(groups) != len(colors) {
		writeError(w, http.StatusBadRequest, "number of colors must match number of muscle groups")
		return
	}
	if err := validateGroups(h.c, groups); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	transparent := boolParam(r, "transparentBackground", false)

	highlights := make([]compositor.Highlight, len(groups))
	for i, g := range groups {
		c, err := compositor.ParseColor(colors[i])
		if err != nil {
			writeError(w, http.StatusBadRequest, "colors["+strconv.Itoa(i)+"]: "+err.Error())
			return
		}
		highlights[i] = compositor.Highlight{Group: g, Color: c}
	}
	renderAndWrite(w, h.c, highlights, transparent)
}

// --- helpers ---

func boolParam(r *http.Request, name string, def bool) bool {
	v := strings.TrimSpace(r.URL.Query().Get(name))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n != 0
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	raw := strings.Split(s, ",")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func validateGroups(c *compositor.Compositor, groups []string) error {
	for _, g := range groups {
		if !c.HasGroup(g) {
			return errors.New("unknown muscle group: " + g)
		}
	}
	return nil
}

func renderAndWrite(w http.ResponseWriter, c *compositor.Compositor, highlights []compositor.Highlight, transparent bool) {
	img, err := c.Render(highlights, transparent)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writePNG(w, img)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func writePNG(w http.ResponseWriter, img image.Image) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		writeError(w, http.StatusInternalServerError, "encode png: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = w.Write(buf.Bytes())
}
