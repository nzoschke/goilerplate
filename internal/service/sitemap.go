package service

import (
	"encoding/xml"
	"log/slog"
	"strings"
	"time"

	"github.com/templui/goilerplate/internal/model"
)

// publicRoutes defines all static public routes that should be included in the sitemap
// Add new public pages here (but not auth-protected pages like /dashboard)
var publicRoutes = []struct {
	Path       string
	Priority   string
	ChangeFreq string
}{
	{"/", "1.0", "daily"},
	{"/blog", "0.8", "daily"},
	{"/docs", "0.8", "weekly"},
	{"/login", "0.3", "monthly"},
	{"/register", "0.3", "monthly"},
	// Add new public routes here, e.g.:
	// {"/about", "0.7", "monthly"},
	// {"/contact", "0.5", "monthly"},
}

type SitemapService struct {
	blogService *BlogService
	docsService *DocsService
	baseURL     string
}

// NewSitemapService creates a new sitemap service
func NewSitemapService(blogService *BlogService, docsService *DocsService, baseURL string) *SitemapService {
	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &SitemapService{
		blogService: blogService,
		docsService: docsService,
		baseURL:     baseURL,
	}
}

// GenerateSitemap generates a complete sitemap including all pages
func (s *SitemapService) GenerateSitemap() ([]byte, error) {
	sitemap := model.Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  []model.SitemapURL{},
	}

	// Add static routes
	staticRoutes := s.getStaticRoutes()
	for _, route := range staticRoutes {
		sitemap.URLs = append(sitemap.URLs, route)
	}

	// Add blog posts
	blogURLs, err := s.getBlogURLs()
	if err != nil {
		// Log error but don't fail - blog might not have posts yet
		slog.Warn("failed to get blog URLs for sitemap", "error", err)
	} else {
		sitemap.URLs = append(sitemap.URLs, blogURLs...)
	}

	// Add documentation pages
	docsURLs := s.getDocsURLs()
	sitemap.URLs = append(sitemap.URLs, docsURLs...)

	// Generate XML
	output, err := xml.MarshalIndent(sitemap, "", "  ")
	if err != nil {
		return nil, err
	}

	// Add XML header
	result := xml.Header + string(output)
	return []byte(result), nil
}

// getStaticRoutes returns the static routes of the application
func (s *SitemapService) getStaticRoutes() []model.SitemapURL {
	today := time.Now().Format("2006-01-02")
	urls := make([]model.SitemapURL, 0, len(publicRoutes))

	for _, route := range publicRoutes {
		urls = append(urls, model.SitemapURL{
			Loc:        s.baseURL + route.Path,
			LastMod:    today,
			ChangeFreq: route.ChangeFreq,
			Priority:   route.Priority,
		})
	}

	return urls
}

// getBlogURLs returns all blog post URLs
func (s *SitemapService) getBlogURLs() ([]model.SitemapURL, error) {
	posts, err := s.blogService.Posts()
	if err != nil {
		return nil, err
	}

	urls := make([]model.SitemapURL, 0, len(posts))
	for _, post := range posts {
		// Use the post date if available, otherwise use today
		lastMod := time.Now().Format("2006-01-02")
		if !post.Date.IsZero() {
			lastMod = post.Date.Format("2006-01-02")
		}

		urls = append(urls, model.SitemapURL{
			Loc:        s.baseURL + "/blog/" + post.Slug,
			LastMod:    lastMod,
			ChangeFreq: "weekly",
			Priority:   "0.7",
		})
	}

	// Also add tag pages for unique tags
	tagMap := make(map[string]bool)
	for _, post := range posts {
		for _, tag := range post.Tags {
			tagMap[tag] = true
		}
	}

	for tag := range tagMap {
		urls = append(urls, model.SitemapURL{
			Loc:        s.baseURL + "/blog/tag/" + tag,
			LastMod:    time.Now().Format("2006-01-02"),
			ChangeFreq: "weekly",
			Priority:   "0.5",
		})
	}

	return urls, nil
}

// getDocsURLs returns all documentation page URLs
func (s *SitemapService) getDocsURLs() []model.SitemapURL {
	// Build docs tree if not already built
	err := s.docsService.BuildDocsTree()
	if err != nil {
		// Log error but don't fail
		slog.Warn("failed to build docs tree for sitemap", "error", err)
		return []model.SitemapURL{}
	}

	pages := s.docsService.FlatDocsList()
	urls := make([]model.SitemapURL, 0, len(pages))

	today := time.Now().Format("2006-01-02")

	for _, page := range pages {
		// Skip pages without content (directory placeholders)
		if page.HTMLContent == "" && len(page.Children) > 0 {
			continue
		}

		// Determine priority based on path depth
		depth := strings.Count(page.Slug, "/")
		priority := "0.6"
		if depth == 0 {
			priority = "0.8" // Top-level docs
		} else if depth == 1 {
			priority = "0.7" // Second-level docs
		}

		urls = append(urls, model.SitemapURL{
			Loc:        s.baseURL + "/docs/" + page.Slug,
			LastMod:    today,
			ChangeFreq: "weekly",
			Priority:   priority,
		})
	}

	return urls
}

