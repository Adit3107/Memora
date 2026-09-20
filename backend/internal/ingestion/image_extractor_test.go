package ingestion

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

type fakeOCRRunner struct {
	text     string
	err      error
	seenPath string
}

func (r *fakeOCRRunner) RunOCR(ctx context.Context, filePath string) (string, error) {
	r.seenPath = filePath
	return r.text, r.err
}

func TestImageExtractorExtractWithOCRText(t *testing.T) {
	runner := &fakeOCRRunner{text: " Hello from OCR \n"}
	extractor := NewImageExtractor(runner)

	got, err := extractor.Extract(context.Background(), ExtractInput{
		FileName:    "note.png",
		ContentType: "image/png",
		Body:        bytes.NewReader(testPNG(t)),
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if got.SourceType != SourceTypeImage {
		t.Fatalf("SourceType = %q, want %q", got.SourceType, SourceTypeImage)
	}
	if got.CleanText != "Hello from OCR" {
		t.Fatalf("CleanText = %q", got.CleanText)
	}
	if got.Metadata["image_format"] != "png" {
		t.Fatalf("image_format = %q", got.Metadata["image_format"])
	}
	if got.Metadata["width"] != "2" || got.Metadata["height"] != "1" {
		t.Fatalf("dimensions = %sx%s", got.Metadata["width"], got.Metadata["height"])
	}
	if runner.seenPath == "" {
		t.Fatal("OCR runner did not receive a file path")
	}
	if _, err := os.Stat(runner.seenPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary OCR file still exists or unexpected stat error: %v", err)
	}
}

func TestImageExtractorExtractFromFilePath(t *testing.T) {
	file, err := os.CreateTemp("", "memora-image-test-*.png")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(testPNG(t)); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	runner := &fakeOCRRunner{text: "File path OCR"}
	extractor := NewImageExtractor(runner)

	got, err := extractor.Extract(context.Background(), ExtractInput{
		FileName:    "photo.png",
		ContentType: "image/png",
		FilePath:    file.Name(),
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if got.CleanText != "File path OCR" {
		t.Fatalf("CleanText = %q", got.CleanText)
	}
	if runner.seenPath == "" {
		t.Fatal("OCR runner did not receive file path")
	}
}

func TestImageExtractorHandlesNoTextAndInvalidImages(t *testing.T) {
	tests := []struct {
		name   string
		input  ExtractInput
		runner *fakeOCRRunner
		err    error
	}{
		{
			name: "image without text",
			input: ExtractInput{
				FileName:    "blank.png",
				ContentType: "image/png",
				Body:        bytes.NewReader(testPNG(t)),
			},
			runner: &fakeOCRRunner{text: "   "},
			err:    ErrEmptyContent,
		},
		{
			name: "unreadable image",
			input: ExtractInput{
				FileName:    "bad.png",
				ContentType: "image/png",
				Body:        bytes.NewReader([]byte("not an image")),
			},
			runner: &fakeOCRRunner{text: "unused"},
			err:    ErrUnsupportedSourceType,
		},
		{
			name: "unsupported file",
			input: ExtractInput{
				FileName:    "archive.zip",
				ContentType: "application/zip",
				Body:        bytes.NewReader([]byte("zip")),
			},
			runner: &fakeOCRRunner{text: "unused"},
			err:    ErrUnsupportedSourceType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewImageExtractor(tt.runner)
			_, err := extractor.Extract(context.Background(), tt.input)
			if !errors.Is(err, tt.err) {
				t.Fatalf("Extract() error = %v, want %v", err, tt.err)
			}
		})
	}
}

func TestImageExtractorReturnsOCRFailure(t *testing.T) {
	wantErr := errors.New("tesseract failed")
	extractor := NewImageExtractor(&fakeOCRRunner{err: wantErr})

	_, err := extractor.Extract(context.Background(), ExtractInput{
		FileName:    "note.png",
		ContentType: "image/png",
		Body:        bytes.NewReader(testPNG(t)),
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Extract() error = %v, want %v", err, wantErr)
	}
}

func testPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.White)
	img.Set(1, 0, color.Black)

	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		t.Fatalf("png.Encode() error = %v", err)
	}

	return buffer.Bytes()
}
