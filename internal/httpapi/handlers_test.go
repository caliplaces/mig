package httpapi_test

import (
	"bytes"
	"encoding/json"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/caliplaces/mig/internal/assets"
	"github.com/caliplaces/mig/internal/compositor"
	"github.com/caliplaces/mig/internal/httpapi"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	lib, err := compositor.NewLibrary(assets.FS)
	if err != nil {
		t.Fatalf("NewLibrary: %v", err)
	}
	ts := httptest.NewServer(httpapi.NewRouter(compositor.New(lib)))
	t.Cleanup(ts.Close)
	return ts
}

func TestGetMuscleGroups(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getMuscleGroups")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	var groups []string
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		t.Fatal(err)
	}
	if len(groups) == 0 {
		t.Fatal("empty groups")
	}
}

func TestGetBaseImage(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getBaseImage")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "image/png" {
		t.Fatalf("content-type=%q", ct)
	}
	if _, err := png.Decode(resp.Body); err != nil {
		t.Fatalf("decode png: %v", err)
	}
}

func TestGetBaseImage_Transparent(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getBaseImage?transparentBackground=1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if _, err := png.Decode(resp.Body); err != nil {
		t.Fatalf("decode png: %v", err)
	}
}

func TestGetImage_Hex(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getImage?muscleGroups=chest,triceps&color=FF0000")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if _, err := png.Decode(resp.Body); err != nil {
		t.Fatalf("decode png: %v", err)
	}
}

func TestGetImage_RGB(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getImage?muscleGroups=chest&color=255,0,0")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestGetImage_MissingGroups(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getImage?color=FF0000")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetImage_UnknownGroup(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getImage?muscleGroups=nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetImage_InvalidColor(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getImage?muscleGroups=chest&color=ZZZ")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetMulticolorImage(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getMulticolorImage?primaryMuscleGroups=chest&secondaryMuscleGroups=triceps&primaryColor=FF0000&secondaryColor=0000FF")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if _, err := png.Decode(resp.Body); err != nil {
		t.Fatalf("decode png: %v", err)
	}
}

func TestGetMulticolorImage_MissingParams(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getMulticolorImage?primaryMuscleGroups=chest&primaryColor=FF0000")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetIndividualColorImage(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getIndividualColorImage?muscleGroups=chest,triceps&colors=FF0000,00FF00")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if _, err := png.Decode(resp.Body); err != nil {
		t.Fatalf("decode png: %v", err)
	}
}

func TestGetIndividualColorImage_MismatchedLengths(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getIndividualColorImage?muscleGroups=chest,triceps&colors=FF0000")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHealthz(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Fatalf("body=%q", string(body))
	}
}

func TestRoot(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestCORSHeader(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/getMuscleGroups")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("ACAO=%q", got)
	}
}

func TestOpenAPISpec(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("openapi:")) {
		t.Fatalf("spec missing openapi key, got first 80 bytes: %q", body[:min(80, len(body))])
	}
}

func TestSwaggerUI(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/docs")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct[:9] != "text/html" {
		t.Fatalf("content-type=%q", ct)
	}
}

func TestUnknownRoute_404(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestMethodNotAllowed_405(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Post(ts.URL+"/getMuscleGroups", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
