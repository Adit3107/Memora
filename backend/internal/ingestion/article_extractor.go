package ingestion

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const defaultUserAgent = "MemoraBot/1.0"

type WebArticleExtractor struct {
	client    HTTPClient
	userAgent string
}

func NewWebArticleExtractor(client HTTPClient) *WebArticleExtractor {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	return &WebArticleExtractor{
		client:    client,
		userAgent: defaultUserAgent,
	}
}

func (e *WebArticleExtractor) Extract(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	detected, err := DetectURL(input.SourceURL)
	if err != nil {
		return IngestionResult{}, err
	}
	if detected != DetectedContentTypeWebArticle {
		return IngestionResult{}, ErrUnsupportedSourceType
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSpace(input.SourceURL), nil)
	if err != nil {
		return IngestionResult{}, ErrInvalidURL
	}
	req.Header.Set("User-Agent", e.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := e.client.Do(req)
	if err != nil {
		return IngestionResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return IngestionResult{}, ErrInaccessibleSource
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return IngestionResult{}, ErrInaccessibleSource
	}
	if !isHTMLResponse(resp.Header.Get("Content-Type")) {
		return IngestionResult{}, ErrUnexpectedContentType
	}

	document, err := html.Parse(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return IngestionResult{}, err
	}

	page := parseHTMLPage(document)
	text := normalizeText(page.MainText)
	if text == "" {
		return IngestionResult{}, ErrEmptyContent
	}

	return IngestionResult{
		SourceType:  SourceTypeArticle,
		Title:       page.Title,
		Description: page.Description,
		Author:      page.Author,
		SourceURL:   input.SourceURL,
		RawText:     page.MainText,
		CleanText:   text,
		Metadata: map[string]string{
			"content_type": string(DetectedContentTypeWebArticle),
		},
	}, nil
}

func isHTMLResponse(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return contentType == "" || strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml+xml")
}

type htmlPage struct {
	Title       string
	Description string
	Author      string
	MainText    string
}

func parseHTMLPage(document *html.Node) htmlPage {
	var page htmlPage
	var articleText []string
	var mainText []string
	var bodyText []string

	var walk func(*html.Node, bool, bool, bool, bool)
	walk = func(node *html.Node, ignored bool, inBody bool, inArticle bool, inMain bool) {
		if node.Type == html.ElementNode {
			tag := strings.ToLower(node.Data)
			if shouldIgnoreElement(tag) {
				ignored = true
			}

			switch tag {
			case "body":
				inBody = true
			case "title":
				if page.Title == "" {
					page.Title = normalizeText(nodeText(node))
				}
			case "meta":
				readMetaNode(node, &page)
			case "article":
				inArticle = true
			case "main":
				inMain = true
			}
		}

		if !ignored && node.Type == html.TextNode {
			text := normalizeText(node.Data)
			if text != "" && inBody {
				bodyText = append(bodyText, text)
				if inArticle {
					articleText = append(articleText, text)
				}
				if inMain {
					mainText = append(mainText, text)
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child, ignored, inBody, inArticle, inMain)
		}
	}
	walk(document, false, false, false, false)

	page.Title = normalizeText(page.Title)
	page.Description = normalizeText(page.Description)
	page.Author = normalizeText(page.Author)
	page.MainText = bestMainText(articleText, mainText, bodyText)

	return page
}

func shouldIgnoreElement(tag string) bool {
	switch tag {
	case "script", "style", "nav", "header", "footer", "aside", "form", "noscript", "svg", "canvas":
		return true
	default:
		return false
	}
}

func readMetaNode(node *html.Node, page *htmlPage) {
	name := strings.ToLower(attrValue(node, "name"))
	property := strings.ToLower(attrValue(node, "property"))
	content := attrValue(node, "content")
	if content == "" {
		return
	}

	switch {
	case page.Description == "" && (name == "description" || property == "og:description"):
		page.Description = content
	case page.Author == "" && (name == "author" || property == "article:author"):
		page.Author = content
	case page.Title == "" && property == "og:title":
		page.Title = content
	}
}

func attrValue(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, key) {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

func nodeText(node *html.Node) string {
	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
			builder.WriteString(" ")
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)

	return builder.String()
}

func bestMainText(articleText []string, mainText []string, bodyText []string) string {
	switch {
	case len(articleText) > 0:
		return joinText(articleText)
	case len(mainText) > 0:
		return joinText(mainText)
	default:
		return joinText(bodyText)
	}
}

var _ Extractor = (*WebArticleExtractor)(nil)
