package ingestion

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestDocumentExtractorExtractTXT(t *testing.T) {
	extractor := NewDocumentExtractor()

	got, err := extractor.Extract(context.Background(), ExtractInput{
		FileName:    "notes.txt",
		ContentType: "text/plain",
		Body:        strings.NewReader("  first line\n\nsecond line  "),
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if got.SourceType != SourceTypeText {
		t.Fatalf("SourceType = %q, want %q", got.SourceType, SourceTypeText)
	}
	if got.CleanText != "first line second line" {
		t.Fatalf("CleanText = %q", got.CleanText)
	}
}

func TestDocumentExtractorExtractCSV(t *testing.T) {
	extractor := NewDocumentExtractor()

	got, err := extractor.Extract(context.Background(), ExtractInput{
		FileName:    "data.csv",
		ContentType: "text/csv",
		Body:        strings.NewReader("name,topic\nAda,Go\nGrace,Compilers\n"),
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if !strings.Contains(got.CleanText, "Row 2") || !strings.Contains(got.CleanText, "Ada") {
		t.Fatalf("CleanText = %q", got.CleanText)
	}
	if got.Metadata["row_count"] != "3" {
		t.Fatalf("row_count = %q", got.Metadata["row_count"])
	}
}

func TestDocumentExtractorExtractDOCX(t *testing.T) {
	body := zipBytes(t, map[string]string{
		"word/document.xml": `<w:document xmlns:w="word"><w:body><w:p><w:r><w:t>Hello DOCX</w:t></w:r></w:p></w:body></w:document>`,
	})
	extractor := NewDocumentExtractor()

	got, err := extractor.Extract(context.Background(), ExtractInput{
		FileName:    "doc.docx",
		ContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if got.CleanText != "Hello DOCX" {
		t.Fatalf("CleanText = %q", got.CleanText)
	}
}

func TestDocumentExtractorExtractPPTX(t *testing.T) {
	body := zipBytes(t, map[string]string{
		"ppt/slides/slide1.xml": `<p:sld xmlns:p="p" xmlns:a="a"><p:cSld><a:t>Slide One</a:t></p:cSld></p:sld>`,
		"ppt/slides/slide2.xml": `<p:sld xmlns:p="p" xmlns:a="a"><p:cSld><a:t>Slide Two</a:t></p:cSld></p:sld>`,
	})
	extractor := NewDocumentExtractor()

	got, err := extractor.Extract(context.Background(), ExtractInput{
		FileName:    "slides.pptx",
		ContentType: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if len(got.Pages) != 2 {
		t.Fatalf("Pages length = %d, want 2", len(got.Pages))
	}
	if !strings.Contains(got.CleanText, "Slide One") || !strings.Contains(got.CleanText, "Slide Two") {
		t.Fatalf("CleanText = %q", got.CleanText)
	}
}

func TestDocumentExtractorExtractXLSX(t *testing.T) {
	body := zipBytes(t, map[string]string{
		"xl/sharedStrings.xml":     `<sst xmlns="spreadsheet"><si><t>Name</t></si><si><t>Ada</t></si></sst>`,
		"xl/worksheets/sheet1.xml": `<worksheet xmlns="spreadsheet"><sheetData><row><c t="s"><v>0</v></c><c t="s"><v>1</v></c><c><v>42</v></c></row></sheetData></worksheet>`,
	})
	extractor := NewDocumentExtractor()

	got, err := extractor.Extract(context.Background(), ExtractInput{
		FileName:    "sheet.xlsx",
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if len(got.Pages) != 1 {
		t.Fatalf("Pages length = %d, want 1", len(got.Pages))
	}
	if !strings.Contains(got.CleanText, "Name") || !strings.Contains(got.CleanText, "Ada") || !strings.Contains(got.CleanText, "42") {
		t.Fatalf("CleanText = %q", got.CleanText)
	}
}

func TestDocumentExtractorExtractPDF(t *testing.T) {
	extractor := NewDocumentExtractor()

	got, err := extractor.Extract(context.Background(), ExtractInput{
		FileName:    "paper.pdf",
		ContentType: "application/pdf",
		Body:        strings.NewReader("%PDF-1.4\n1 0 obj << /Type /Page >> stream\nBT (Hello PDF text) Tj ET\nendstream"),
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if !strings.Contains(got.CleanText, "Hello PDF text") {
		t.Fatalf("CleanText = %q", got.CleanText)
	}
	if got.Metadata["page_count"] != "1" {
		t.Fatalf("page_count = %q", got.Metadata["page_count"])
	}
}

func TestDocumentExtractorHandlesInvalidAndEmptyFiles(t *testing.T) {
	tests := []struct {
		name  string
		input ExtractInput
		err   error
	}{
		{
			name: "empty body",
			input: ExtractInput{
				FileName:    "notes.txt",
				ContentType: "text/plain",
				Body:        strings.NewReader(""),
			},
			err: ErrEmptyContent,
		},
		{
			name: "unsupported file",
			input: ExtractInput{
				FileName:    "archive.zip",
				ContentType: "application/zip",
				Body:        strings.NewReader("zip"),
			},
			err: ErrUnsupportedSourceType,
		},
		{
			name: "invalid docx",
			input: ExtractInput{
				FileName:    "bad.docx",
				ContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				Body:        strings.NewReader("not a zip"),
			},
		},
	}

	extractor := NewDocumentExtractor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractor.Extract(context.Background(), tt.input)
			if tt.err != nil && !errors.Is(err, tt.err) {
				t.Fatalf("Extract() error = %v, want %v", err, tt.err)
			}
			if tt.err == nil && err == nil {
				t.Fatal("Extract() error = nil, want error")
			}
		})
	}
}

func zipBytes(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%q) error = %v", name, err)
		}
		if _, err := file.Write([]byte(content)); err != nil {
			t.Fatalf("Write(%q) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	return buffer.Bytes()
}
