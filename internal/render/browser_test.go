package render

import (
	"context"
	"testing"
	"time"
)

func TestFindChromium(t *testing.T) {
	path := FindChromium()
	// Should not panic, and returns either a valid path or empty string
	t.Logf("Detected Chromium path: %q", path)
}

func TestRendererAvailability(t *testing.T) {
	r := NewRenderer("")
	t.Logf("Renderer available: %v, browser path: %s", r.Available, r.BrowserPath)

	if !r.Available {
		_, err := r.Render(context.Background(), "https://example.com")
		if err != ErrRenderNotSupported {
			t.Fatalf("expected ErrRenderNotSupported when not available, got %v", err)
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		snap, err := r.Render(ctx, "https://example.com")
		if err != nil {
			t.Logf("Render failed (expected in network-isolated test): %v", err)
		} else {
			if snap.Title == "" {
				t.Errorf("expected title in rendered snapshot")
			}
			t.Logf("Successfully rendered: title=%q textLen=%d", snap.Title, snap.TextLength)
		}
	}
}
