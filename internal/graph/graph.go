package graph

import "sync"

// LinkGraph builds an adjacency structure of internal links across crawled pages.
type LinkGraph struct {
	mu       sync.RWMutex
	inbound  map[string]map[string]bool // target → set of source URLs
	outbound map[string]map[string]bool // source → set of target URLs
	pages    map[string]bool            // all known crawled pages
}

// New creates an empty LinkGraph.
func New() *LinkGraph {
	return &LinkGraph{
		inbound:  make(map[string]map[string]bool),
		outbound: make(map[string]map[string]bool),
		pages:    make(map[string]bool),
	}
}

// AddPage registers a crawled page and its outgoing links.
func (g *LinkGraph) AddPage(pageURL string, outLinks []string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.pages[pageURL] = true

	if g.outbound[pageURL] == nil {
		g.outbound[pageURL] = make(map[string]bool)
	}
	for _, link := range outLinks {
		g.outbound[pageURL][link] = true

		if g.inbound[link] == nil {
			g.inbound[link] = make(map[string]bool)
		}
		g.inbound[link][pageURL] = true
	}
}

// Orphans returns pages that were crawled but have no inbound internal links
// from other crawled pages. The seed URL is excluded since it naturally has
// no inbound link within the crawl.
func (g *LinkGraph) Orphans(seedURL string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var orphans []string
	for page := range g.pages {
		if page == seedURL {
			continue
		}
		inbound := g.inbound[page]
		// Count only inbound links from other crawled pages
		count := 0
		for src := range inbound {
			if g.pages[src] {
				count++
			}
		}
		if count == 0 {
			orphans = append(orphans, page)
		}
	}
	return orphans
}

// Depth computes the shortest crawl depth from the seed URL using BFS.
// Returns a map of URL → depth. The seed is depth 0.
func (g *LinkGraph) Depth(seedURL string) map[string]int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	depths := map[string]int{seedURL: 0}
	queue := []string{seedURL}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		currentDepth := depths[current]

		for target := range g.outbound[current] {
			if _, seen := depths[target]; !seen {
				depths[target] = currentDepth + 1
				queue = append(queue, target)
			}
		}
	}
	return depths
}

// InboundCount returns the number of internal inbound links to a URL
// from other crawled pages.
func (g *LinkGraph) InboundCount(url string) int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	count := 0
	for src := range g.inbound[url] {
		if g.pages[src] {
			count++
		}
	}
	return count
}

// OutboundCount returns the number of outgoing internal links from a URL.
func (g *LinkGraph) OutboundCount(url string) int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.outbound[url])
}

// PageCount returns the total number of crawled pages.
func (g *LinkGraph) PageCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.pages)
}

// Pages returns all crawled page URLs.
func (g *LinkGraph) Pages() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	pages := make([]string, 0, len(g.pages))
	for p := range g.pages {
		pages = append(pages, p)
	}
	return pages
}
