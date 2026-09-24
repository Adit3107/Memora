package ingestion

type SourceType string

const (
	SourceTypeArticle  SourceType = "article"
	SourceTypeVideo    SourceType = "video"
	SourceTypeReel     SourceType = "reel"
	SourceTypeDocument SourceType = "document"
	SourceTypeImage    SourceType = "image"
	SourceTypeText     SourceType = "text"
)

type DetectedContentType string

const (
	DetectedContentTypeYouTube       DetectedContentType = "youtube"
	DetectedContentTypeInstagramReel DetectedContentType = "instagram"
	DetectedContentTypeFacebookReel  DetectedContentType = "facebook"
	DetectedContentTypeWebArticle    DetectedContentType = "web_article"
	DetectedContentTypePDF           DetectedContentType = "pdf"
	DetectedContentTypeDOCX          DetectedContentType = "docx"
	DetectedContentTypePPTX          DetectedContentType = "pptx"
	DetectedContentTypeTXT           DetectedContentType = "txt"
	DetectedContentTypeCSV           DetectedContentType = "csv"
	DetectedContentTypeXLSX          DetectedContentType = "xlsx"
	DetectedContentTypeImage         DetectedContentType = "image"
)

type IngestionResult struct {
	SourceType SourceType `json:"source_type"`

	Title       string `json:"title"`
	Description string `json:"description"`
	Author      string `json:"author,omitempty"`
	SourceURL   string `json:"source_url,omitempty"`

	RawText   string `json:"raw_text"`
	CleanText string `json:"clean_text"`

	Transcript []TranscriptSegment `json:"transcript,omitempty"`
	Pages      []DocumentPage      `json:"pages,omitempty"`
	Metadata   map[string]string   `json:"metadata,omitempty"`
}

type TranscriptSegment struct {
	StartSeconds float64 `json:"start_seconds"`
	EndSeconds   float64 `json:"end_seconds"`
	Text         string  `json:"text"`
}

type DocumentPage struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

// Why this file exists:
// Ingestion converts many source formats into one common shape.
// Later phases can chunk, embed, and search this result without knowing
// whether the original source was a web page, video, document, image, or text file.
