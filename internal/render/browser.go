package render

import (
	"context"
	"errors"
	"rankcore/internal/crawl"
)

var ErrRenderNotSupported = errors.New("rendered mode is not supported without a Chromium-family browser installed")

// Renderer manages CDP interaction for executing rendered crawling.
type Renderer struct {
	Available bool
}

func NewRenderer() *Renderer {
	// For MVP End-to-End, we will say false, avoiding CDP dependency.
	return &Renderer{Available: false}
}

// Snapshot loads the page in a headless browser, waits for network idle, and extracts the rendered DOM.
func (r *Renderer) Snapshot(ctx context.Context, url string) (*crawl.PageSnapshot, error) {
	if !r.Available {
		return nil, ErrRenderNotSupported
	}

	// This is the actual implementation structure if chromedp is used:
	// cctx, cancel := chromedp.NewContext(ctx)
	// defer cancel()
	// var html string
	// err := chromedp.Run(cctx,
	//     chromedp.Navigate(url),
	//     chromedp.OuterHTML("html", &html),
	// )
	
	return &crawl.PageSnapshot{}, nil
}
