package ingestion

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type RedditExtractor struct {
	client      HTTPClient
	jsonBaseURL string
	userAgent   string
}

type RedditExtractorOption func(*RedditExtractor)

func NewRedditExtractor(client HTTPClient, options ...RedditExtractorOption) *RedditExtractor {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	extractor := &RedditExtractor{
		client:      client,
		jsonBaseURL: "",
		userAgent:   defaultUserAgent,
	}

	for _, option := range options {
		option(extractor)
	}

	return extractor
}

func WithRedditJSONBaseURL(baseURL string) RedditExtractorOption {
	return func(extractor *RedditExtractor) {
		extractor.jsonBaseURL = strings.TrimRight(baseURL, "/")
	}
}

func (e *RedditExtractor) Extract(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	detected, err := DetectURL(input.SourceURL)
	if err != nil {
		return IngestionResult{}, err
	}
	if detected != DetectedContentTypeReddit {
		return IngestionResult{}, ErrUnsupportedSourceType
	}

	endpoint, err := e.redditJSONURL(input.SourceURL)
	if err != nil {
		return IngestionResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return IngestionResult{}, ErrInvalidURL
	}
	req.Header.Set("User-Agent", e.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return IngestionResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
		return IngestionResult{}, ErrInaccessibleSource
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return IngestionResult{}, ErrInaccessibleSource
	}

	var listing redditListing
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2*1024*1024)).Decode(&listing); err != nil {
		return IngestionResult{}, err
	}

	post, err := firstRedditPost(listing)
	if err != nil {
		return IngestionResult{}, err
	}

	text := joinText([]string{post.Title, post.SelfText})
	if text == "" {
		return IngestionResult{}, ErrEmptyContent
	}

	metadata := map[string]string{
		"content_type": string(DetectedContentTypeReddit),
		"provider":     "reddit",
		"subreddit":    post.Subreddit,
		"permalink":    post.Permalink,
	}
	if post.CreatedUTC > 0 {
		metadata["created_utc"] = time.Unix(int64(post.CreatedUTC), 0).UTC().Format(time.RFC3339)
	}
	if post.Score != 0 {
		metadata["score"] = intString(post.Score)
	}
	if post.NumComments != 0 {
		metadata["num_comments"] = intString(post.NumComments)
	}

	return IngestionResult{
		SourceType:  SourceTypeArticle,
		Title:       normalizeText(post.Title),
		Author:      normalizeText(post.Author),
		SourceURL:   input.SourceURL,
		RawText:     text,
		CleanText:   normalizeText(text),
		Description: normalizeText(post.SelfText),
		Metadata:    metadata,
	}, nil
}

func (e *RedditExtractor) redditJSONURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", ErrInvalidURL
	}

	path := strings.TrimRight(parsedURL.Path, "/")
	if !strings.HasSuffix(path, ".json") {
		path += ".json"
	}

	if e.jsonBaseURL != "" {
		baseURL, err := url.Parse(e.jsonBaseURL)
		if err != nil {
			return "", err
		}
		baseURL.Path = strings.TrimRight(baseURL.Path, "/") + path
		baseURL.RawQuery = ""
		baseURL.Fragment = ""
		return baseURL.String(), nil
	}

	parsedURL.Path = path
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""

	return parsedURL.String(), nil
}

func firstRedditPost(listing redditListing) (redditPostData, error) {
	if len(listing) == 0 || len(listing[0].Data.Children) == 0 {
		return redditPostData{}, ErrEmptyContent
	}

	post := listing[0].Data.Children[0].Data
	if normalizeText(post.Title) == "" && normalizeText(post.SelfText) == "" {
		return redditPostData{}, ErrEmptyContent
	}

	return post, nil
}

type redditListing []struct {
	Data struct {
		Children []struct {
			Data redditPostData `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type redditPostData struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Subreddit   string  `json:"subreddit"`
	Permalink   string  `json:"permalink"`
	SelfText    string  `json:"selftext"`
	CreatedUTC  float64 `json:"created_utc"`
	Score       int     `json:"score"`
	NumComments int     `json:"num_comments"`
}

func intString(value int) string {
	return strconv.Itoa(value)
}

var _ Extractor = (*RedditExtractor)(nil)
