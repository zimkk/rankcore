package render

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/zimkk/rankcore/internal/crawl"
	"github.com/zimkk/rankcore/internal/extract"
)

var ErrRenderNotSupported = errors.New("rendered mode is not supported without a Chromium-family browser installed")

// Renderer manages headless browser execution to extract client-side rendered DOM.
type Renderer struct {
	BrowserPath string
	Available   bool
	Timeout     time.Duration
}

// FindChromium detects an installed Chromium-family browser (Chrome, Chromium, Edge, Brave).
func FindChromium() string {
	var candidates []string
	switch runtime.GOOS {
	case "darwin":
		candidates = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
		}
	case "linux":
		candidates = []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		programFiles := os.Getenv("ProgramFiles")
		programFilesX86 := os.Getenv("ProgramFiles(x86)")
		candidates = []string{
			programFiles + `\Google\Chrome\Application\chrome.exe`,
			programFilesX86 + `\Google\Chrome\Application\chrome.exe`,
			localAppData + `\Google\Chrome\Application\chrome.exe`,
			programFiles + `\Microsoft\Edge\Application\msedge.exe`,
			programFilesX86 + `\Microsoft\Edge\Application\msedge.exe`,
		}
	}

	for _, p := range candidates {
		if p != "" {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	// Try PATH lookups
	pathCandidates := []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "chrome", "msedge"}
	for _, name := range pathCandidates {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}

	return ""
}

// NewRenderer creates a Renderer, discovering Chromium automatically if path is empty.
func NewRenderer(browserPath string) *Renderer {
	if browserPath == "" {
		browserPath = FindChromium()
	}

	return &Renderer{
		BrowserPath: browserPath,
		Available:   browserPath != "",
		Timeout:     15 * time.Second,
	}
}

// Render loads the page in headless Chromium and captures the rendered DOM.
func (r *Renderer) Render(ctx context.Context, pageURL string) (*crawl.RenderedSnapshot, error) {
	if !r.Available {
		return nil, ErrRenderNotSupported
	}

	renderCtx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()

	// Use modern headless Chromium with --dump-dom
	args := []string{
		"--headless=new",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--dump-dom",
		pageURL,
	}

	cmd := exec.CommandContext(renderCtx, r.BrowserPath, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("browser render failed for %s: %w", pageURL, err)
	}

	renderedHTML := stdout.Bytes()
	if len(renderedHTML) == 0 {
		return nil, fmt.Errorf("browser returned empty DOM for %s", pageURL)
	}

	meta, err := extract.Extract(bytes.NewReader(renderedHTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse rendered DOM: %w", err)
	}

	return &crawl.RenderedSnapshot{
		Title:         meta.Title,
		Canonical:     meta.Canonical,
		RobotsMeta:    meta.RobotsMeta,
		InternalLinks: meta.InternalLinks,
		TextLength:    meta.TextLength,
	}, nil
}
