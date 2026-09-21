package ingestion

import (
	"mime"
	"net"
	"net/url"
	"path/filepath"
	"strings"
)

func DetectURL(rawURL string) (DetectedContentType, error) {
	parsedURL, err := url.Parse(NormalizeSourceURL(rawURL))
	if err != nil {
		return "", ErrInvalidURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", ErrInvalidURL
	}
	if parsedURL.Host == "" {
		return "", ErrInvalidURL
	}

	host := normalizedHost(parsedURL.Host)
	switch {
	case isYouTubeHost(host):
		return DetectedContentTypeYouTube, nil
	default:
		return DetectedContentTypeWebArticle, nil
	}
}

func NormalizeSourceURL(rawURL string) string {
	value := strings.TrimSpace(rawURL)
	if strings.HasPrefix(value, "<") && strings.HasSuffix(value, ">") {
		value = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "<"), ">"))
	}

	if strings.HasPrefix(value, "[") {
		labelEnd := strings.Index(value, "](")
		if labelEnd > 0 && strings.HasSuffix(value, ")") {
			value = strings.TrimSpace(value[labelEnd+2 : len(value)-1])
		}
	}

	return value
}

func DetectFile(fileName string, mimeType string) (DetectedContentType, error) {
	mediaType := normalizedMediaType(mimeType)
	if detected, ok := detectByMIMEType(mediaType); ok {
		return detected, nil
	}

	extension := strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
	if detected, ok := detectByExtension(extension); ok {
		return detected, nil
	}

	return "", ErrUnsupportedSourceType
}

func normalizedHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if withoutPort, _, err := net.SplitHostPort(host); err == nil {
		host = withoutPort
	}
	return strings.TrimPrefix(host, "www.")
}

func normalizedMediaType(mimeType string) string {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(strings.ToLower(mimeType)))
	if err != nil {
		return strings.TrimSpace(strings.ToLower(mimeType))
	}
	return mediaType
}

func isYouTubeHost(host string) bool {
	return host == "youtube.com" || host == "m.youtube.com" || host == "youtu.be"
}

func detectByMIMEType(mediaType string) (DetectedContentType, bool) {
	switch {
	case mediaType == "application/pdf":
		return DetectedContentTypePDF, true
	case mediaType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return DetectedContentTypeDOCX, true
	case mediaType == "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return DetectedContentTypePPTX, true
	case mediaType == "text/plain":
		return DetectedContentTypeTXT, true
	case mediaType == "text/csv" || mediaType == "application/csv":
		return DetectedContentTypeCSV, true
	case mediaType == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return DetectedContentTypeXLSX, true
	case strings.HasPrefix(mediaType, "image/"):
		return DetectedContentTypeImage, true
	default:
		return "", false
	}
}

func detectByExtension(extension string) (DetectedContentType, bool) {
	switch extension {
	case ".pdf":
		return DetectedContentTypePDF, true
	case ".docx":
		return DetectedContentTypeDOCX, true
	case ".pptx":
		return DetectedContentTypePPTX, true
	case ".txt":
		return DetectedContentTypeTXT, true
	case ".csv":
		return DetectedContentTypeCSV, true
	case ".xlsx":
		return DetectedContentTypeXLSX, true
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tif", ".tiff":
		return DetectedContentTypeImage, true
	default:
		return "", false
	}
}

// Why this file exists:
// Detection chooses the correct ingestion path before extraction begins.
// URLs are parsed with net/url so malformed input is rejected early, and files
// are checked by MIME type first with file extension as a practical fallback.
