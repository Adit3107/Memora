package ingestion

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type OCRRunner interface {
	RunOCR(ctx context.Context, filePath string) (string, error)
}

type TesseractRunner struct {
	Binary string
}

func (r TesseractRunner) RunOCR(ctx context.Context, filePath string) (string, error) {
	binary := strings.TrimSpace(r.Binary)
	if binary == "" {
		binary = "tesseract"
	}

	command := exec.CommandContext(ctx, binary, filePath, "stdout")
	output, err := command.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

type ImageExtractor struct {
	runner OCRRunner
}

func NewImageExtractor(runner OCRRunner) *ImageExtractor {
	if runner == nil {
		runner = TesseractRunner{}
	}

	return &ImageExtractor{runner: runner}
}

func (e *ImageExtractor) Extract(ctx context.Context, input ExtractInput) (IngestionResult, error) {
	if err := ctx.Err(); err != nil {
		return IngestionResult{}, err
	}

	detected, err := DetectFile(input.FileName, input.ContentType)
	if err != nil {
		return IngestionResult{}, err
	}
	if detected != DetectedContentTypeImage {
		return IngestionResult{}, ErrUnsupportedSourceType
	}

	imageData, sourcePath, cleanup, err := imageDataAndPath(input)
	if err != nil {
		return IngestionResult{}, err
	}
	if cleanup != nil {
		defer cleanup()
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err != nil {
		return IngestionResult{}, ErrUnsupportedSourceType
	}

	ocrText, err := e.runner.RunOCR(ctx, sourcePath)
	if err != nil {
		return IngestionResult{}, err
	}

	cleanText := normalizeText(ocrText)
	if cleanText == "" {
		return IngestionResult{}, ErrEmptyContent
	}

	return IngestionResult{
		SourceType: SourceTypeImage,
		Title:      strings.TrimSpace(input.FileName),
		RawText:    ocrText,
		CleanText:  cleanText,
		Metadata: map[string]string{
			"content_type": string(DetectedContentTypeImage),
			"file_name":    strings.TrimSpace(input.FileName),
			"image_format": format,
			"width":        fmt.Sprintf("%d", config.Width),
			"height":       fmt.Sprintf("%d", config.Height),
		},
	}, nil
}

func imageDataAndPath(input ExtractInput) ([]byte, string, func(), error) {
	if input.Body != nil {
		data, err := io.ReadAll(io.LimitReader(input.Body, maxDocumentBytes+1))
		if err != nil {
			return nil, "", nil, err
		}
		if len(data) == 0 {
			return nil, "", nil, ErrEmptyContent
		}
		if len(data) > maxDocumentBytes {
			return nil, "", nil, ErrInaccessibleSource
		}

		tempFile, err := os.CreateTemp("", "memora-ocr-*"+safeImageExtension(input.FileName))
		if err != nil {
			return nil, "", nil, err
		}
		tempPath := tempFile.Name()
		cleanup := func() {
			_ = os.Remove(tempPath)
		}

		if _, err := tempFile.Write(data); err != nil {
			_ = tempFile.Close()
			cleanup()
			return nil, "", nil, err
		}
		if err := tempFile.Close(); err != nil {
			cleanup()
			return nil, "", nil, err
		}

		return data, tempPath, cleanup, nil
	}

	filePath, err := validateImagePath(input.FilePath)
	if err != nil {
		return nil, "", nil, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", nil, err
	}
	if len(data) == 0 {
		return nil, "", nil, ErrEmptyContent
	}

	return data, filePath, nil, nil
}

func validateImagePath(filePath string) (string, error) {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return "", ErrEmptyContent
	}

	cleanPath, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		return "", err
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", ErrUnsupportedSourceType
	}

	return cleanPath, nil
}

func safeImageExtension(fileName string) string {
	extension := strings.ToLower(filepath.Ext(fileName))
	switch extension {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tif", ".tiff", ".webp":
		return extension
	default:
		return ".img"
	}
}

var _ Extractor = (*ImageExtractor)(nil)
var _ OCRRunner = TesseractRunner{}

// Why this file exists:
// Images need OCR before they can become searchable text.
// This extractor validates image data, records simple image metadata, and calls
// Tesseract through a fixed command boundary instead of executing user input.
