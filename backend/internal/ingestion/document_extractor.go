package ingestion

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"
)

const maxDocumentBytes = 50 * 1024 * 1024

type DocumentExtractor struct{}

func NewDocumentExtractor() *DocumentExtractor {
	return &DocumentExtractor{}
}

func (e *DocumentExtractor) Extract(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	if err := ctx.Err(); err != nil {
		return IngestionResult{}, err
	}
	if input.Body == nil {
		return IngestionResult{}, ErrEmptyContent
	}

	detected, err := DetectFile(input.FileName, input.ContentType)
	if err != nil {
		return IngestionResult{}, err
	}

	data, err := io.ReadAll(io.LimitReader(input.Body, maxDocumentBytes+1))
	if err != nil {
		return IngestionResult{}, err
	}
	if len(data) == 0 {
		return IngestionResult{}, ErrEmptyContent
	}
	if len(data) > maxDocumentBytes {
		return IngestionResult{}, ErrInaccessibleSource
	}

	switch detected {
	case DetectedContentTypePDF:
		return buildDocumentResult(input, detected, extractPDFText(data))
	case DetectedContentTypeDOCX:
		extracted, err := extractDOCXText(data)
		if err != nil {
			return IngestionResult{}, err
		}
		return buildDocumentResult(input, detected, extracted)
	case DetectedContentTypePPTX:
		extracted, err := extractPPTXText(data)
		if err != nil {
			return IngestionResult{}, err
		}
		return buildDocumentResult(input, detected, extracted)
	case DetectedContentTypeTXT:
		text := string(data)
		return buildDocumentResult(input, detected, extractedDocument{Text: text})
	case DetectedContentTypeCSV:
		extracted, err := extractCSVText(data)
		if err != nil {
			return IngestionResult{}, err
		}
		return buildDocumentResult(input, detected, extracted)
	case DetectedContentTypeXLSX:
		extracted, err := extractXLSXText(data)
		if err != nil {
			return IngestionResult{}, err
		}
		return buildDocumentResult(input, detected, extracted)
	default:
		return IngestionResult{}, ErrUnsupportedSourceType
	}
}

type extractedDocument struct {
	Text     string
	Pages    []DocumentPage
	Metadata map[string]string
}

func buildDocumentResult(input ExtractInput, detected DetectedContentType, extracted extractedDocument) (IngestionResult, error) {
	cleanText := normalizeText(extracted.Text)
	if cleanText == "" {
		return IngestionResult{}, ErrEmptyContent
	}

	sourceType := SourceTypeDocument
	if detected == DetectedContentTypeTXT {
		sourceType = SourceTypeText
	}

	metadata := map[string]string{
		"content_type": string(detected),
		"file_name":    strings.TrimSpace(input.FileName),
	}
	for key, value := range extracted.Metadata {
		if normalizeText(value) != "" {
			metadata[key] = value
		}
	}

	return IngestionResult{
		SourceType: sourceType,
		Title:      strings.TrimSpace(input.FileName),
		RawText:    extracted.Text,
		CleanText:  cleanText,
		Pages:      extracted.Pages,
		Metadata:   metadata,
	}, nil
}

func extractPDFText(data []byte) extractedDocument {
	text := string(data)
	pageCount := strings.Count(text, "/Type /Page")
	if pageCount == 0 {
		pageCount = strings.Count(text, "/Type/Page")
	}

	decoded := decodePDFLiteralStrings(text)
	pages := []DocumentPage(nil)
	if pageCount > 0 && decoded != "" {
		pages = []DocumentPage{{Index: 1, Text: decoded}}
	}

	return extractedDocument{
		Text:  decoded,
		Pages: pages,
		Metadata: map[string]string{
			"page_count": fmt.Sprintf("%d", pageCount),
		},
	}
}

func decodePDFLiteralStrings(text string) string {
	matches := regexp.MustCompile(`\(([^()]*)\)`).FindAllStringSubmatch(text, -1)
	parts := make([]string, 0, len(matches))
	for _, match := range matches {
		value := strings.NewReplacer(`\(`, "(", `\)`, ")", `\\`, `\`).Replace(match[1])
		value = normalizeText(value)
		if value != "" && !strings.HasPrefix(value, "/") {
			parts = append(parts, value)
		}
	}
	return joinText(parts)
}

func extractDOCXText(data []byte) (extractedDocument, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return extractedDocument{}, err
	}

	file := zipFile(reader, "word/document.xml")
	if file == nil {
		return extractedDocument{}, ErrUnexpectedContentType
	}

	parts, err := xmlTextsFromZipFile(file, map[string]bool{"t": true})
	if err != nil {
		return extractedDocument{}, err
	}

	return extractedDocument{Text: joinText(parts)}, nil
}

func extractPPTXText(data []byte) (extractedDocument, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return extractedDocument{}, err
	}

	slideFiles := make([]*zip.File, 0)
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
			slideFiles = append(slideFiles, file)
		}
	}
	sort.Slice(slideFiles, func(i, j int) bool {
		return slideFiles[i].Name < slideFiles[j].Name
	})

	allParts := make([]string, 0)
	pages := make([]DocumentPage, 0, len(slideFiles))
	for index, file := range slideFiles {
		parts, err := xmlTextsFromZipFile(file, map[string]bool{"t": true})
		if err != nil {
			return extractedDocument{}, err
		}
		text := joinText(parts)
		if text == "" {
			continue
		}
		allParts = append(allParts, fmt.Sprintf("Slide %d: %s", index+1, text))
		pages = append(pages, DocumentPage{Index: index + 1, Text: text})
	}

	return extractedDocument{
		Text:  joinText(allParts),
		Pages: pages,
		Metadata: map[string]string{
			"slide_count": fmt.Sprintf("%d", len(slideFiles)),
		},
	}, nil
}

func extractCSVText(data []byte) (extractedDocument, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return extractedDocument{}, err
	}

	lines := make([]string, 0, len(records))
	for index, record := range records {
		cells := make([]string, 0, len(record))
		for column, value := range record {
			value = normalizeText(value)
			if value != "" {
				cells = append(cells, fmt.Sprintf("Column %d: %s", column+1, value))
			}
		}
		if len(cells) > 0 {
			lines = append(lines, fmt.Sprintf("Row %d: %s", index+1, strings.Join(cells, "; ")))
		}
	}

	return extractedDocument{
		Text: joinText(lines),
		Metadata: map[string]string{
			"row_count": fmt.Sprintf("%d", len(records)),
		},
	}, nil
}

func extractXLSXText(data []byte) (extractedDocument, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return extractedDocument{}, err
	}

	sharedStrings, err := readXLSXSharedStrings(reader)
	if err != nil {
		return extractedDocument{}, err
	}

	sheetFiles := make([]*zip.File, 0)
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "xl/worksheets/sheet") && strings.HasSuffix(file.Name, ".xml") {
			sheetFiles = append(sheetFiles, file)
		}
	}
	sort.Slice(sheetFiles, func(i, j int) bool {
		return sheetFiles[i].Name < sheetFiles[j].Name
	})

	parts := make([]string, 0)
	pages := make([]DocumentPage, 0, len(sheetFiles))
	for index, file := range sheetFiles {
		sheetText, err := readXLSXSheet(file, sharedStrings)
		if err != nil {
			return extractedDocument{}, err
		}
		if sheetText == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("Sheet %d: %s", index+1, sheetText))
		pages = append(pages, DocumentPage{Index: index + 1, Text: sheetText})
	}

	return extractedDocument{
		Text:  joinText(parts),
		Pages: pages,
		Metadata: map[string]string{
			"sheet_count": fmt.Sprintf("%d", len(sheetFiles)),
		},
	}, nil
}

func zipFile(reader *zip.Reader, name string) *zip.File {
	cleanName := path.Clean(name)
	for _, file := range reader.File {
		if path.Clean(file.Name) == cleanName {
			return file
		}
	}
	return nil
}

func xmlTextsFromZipFile(file *zip.File, names map[string]bool) ([]string, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	parts := make([]string, 0)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		start, ok := token.(xml.StartElement)
		if !ok || !names[start.Name.Local] {
			continue
		}

		var value string
		if err := decoder.DecodeElement(&value, &start); err != nil {
			return nil, err
		}
		if normalized := normalizeText(value); normalized != "" {
			parts = append(parts, normalized)
		}
	}

	return parts, nil
}

func readXLSXSharedStrings(reader *zip.Reader) ([]string, error) {
	file := zipFile(reader, "xl/sharedStrings.xml")
	if file == nil {
		return nil, nil
	}

	return xmlTextsFromZipFile(file, map[string]bool{"t": true})
}

func readXLSXSheet(file *zip.File, sharedStrings []string) (string, error) {
	rc, err := file.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	parts := make([]string, 0)
	var cellType string

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		switch start.Name.Local {
		case "c":
			cellType = attrByLocalName(start, "t")
		case "v", "t":
			var value string
			if err := decoder.DecodeElement(&value, &start); err != nil {
				return "", err
			}
			value = normalizeText(value)
			if value == "" {
				continue
			}
			if start.Name.Local == "v" && cellType == "s" {
				value = sharedStringValue(sharedStrings, value)
			}
			if value != "" {
				parts = append(parts, value)
			}
		}
	}

	return joinText(parts), nil
}

func attrByLocalName(start xml.StartElement, name string) string {
	for _, attr := range start.Attr {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

func sharedStringValue(sharedStrings []string, rawIndex string) string {
	var index int
	if _, err := fmt.Sscanf(rawIndex, "%d", &index); err != nil {
		return rawIndex
	}
	if index < 0 || index >= len(sharedStrings) {
		return rawIndex
	}
	return sharedStrings[index]
}

var _ Extractor = (*DocumentExtractor)(nil)

// Why this file exists:
// Uploaded documents are binary streams, but Memora needs searchable text.
// This Go extractor is a basic local path for simple formats and tests. The
// target architecture can replace richer PDF/DOCX/PPTX extraction with a Go
// client that calls the internal Python service while keeping this same
// Extractor contract.
