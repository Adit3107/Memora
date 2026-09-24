package ingestion

import "strings"

type ChunkConfig struct {
	MaxCharacters     int
	OverlapCharacters int
}

type ContentChunk struct {
	Index          int               `json:"index"`
	Text           string            `json:"text"`
	SourceType     SourceType        `json:"source_type"`
	PageIndex      *int              `json:"page_index,omitempty"`
	StartSeconds   *float64          `json:"start_seconds,omitempty"`
	EndSeconds     *float64          `json:"end_seconds,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	Embedding      []float64         `json:"-"`
	EmbeddingModel string            `json:"embedding_model,omitempty"`
}

func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		MaxCharacters:     1200,
		OverlapCharacters: 150,
	}
}

func ChunkIngestionResult(result IngestionResult, config ChunkConfig) ([]ContentChunk, error) {
	if config.MaxCharacters <= 0 {
		config.MaxCharacters = DefaultChunkConfig().MaxCharacters
	}
	if config.OverlapCharacters < 0 {
		config.OverlapCharacters = 0
	}
	if config.OverlapCharacters >= config.MaxCharacters {
		config.OverlapCharacters = config.MaxCharacters / 4
	}

	cleaned, err := CleanIngestionResult(result)
	if err != nil {
		return nil, err
	}

	switch {
	case len(cleaned.Transcript) > 0:
		return chunkTranscript(cleaned, config), nil
	case len(cleaned.Pages) > 0:
		return chunkPages(cleaned, config), nil
	default:
		chunks := chunkText(cleaned.CleanText, config)
		if len(chunks) == 0 {
			return nil, ErrEmptyContent
		}
		return withChunkMetadata(chunks, cleaned.SourceType, nil), nil
	}
}

func chunkTranscript(result IngestionResult, config ChunkConfig) []ContentChunk {
	chunks := make([]ContentChunk, 0)
	var current []TranscriptSegment
	var currentText []string
	currentLength := 0

	flush := func() {
		if len(current) == 0 {
			return
		}

		start := current[0].StartSeconds
		end := current[len(current)-1].EndSeconds
		chunks = append(chunks, ContentChunk{
			Index:        len(chunks),
			Text:         joinText(currentText),
			SourceType:   result.SourceType,
			StartSeconds: &start,
			EndSeconds:   &end,
			Metadata:     copyMetadata(result.Metadata),
		})

		current = nil
		currentText = nil
		currentLength = 0
	}

	for _, segment := range result.Transcript {
		text := cleanInlineText(segment.Text)
		if text == "" {
			continue
		}
		if currentLength > 0 && currentLength+len(text)+1 > config.MaxCharacters {
			flush()
		}
		current = append(current, segment)
		currentText = append(currentText, text)
		currentLength += len(text) + 1
	}
	flush()

	if result.SourceType == SourceTypeReel && len(chunks) > 1 {
		fullText := result.CleanText
		if fullText != "" {
			start := result.Transcript[0].StartSeconds
			end := result.Transcript[len(result.Transcript)-1].EndSeconds
			fullMetadata := copyMetadata(result.Metadata)
			if fullMetadata == nil {
				fullMetadata = make(map[string]string)
			}
			fullMetadata["chunk_scope"] = "full_reel"
			fullChunk := ContentChunk{
				Index:        0,
				Text:         fullText,
				SourceType:   result.SourceType,
				StartSeconds: &start,
				EndSeconds:   &end,
				Metadata:     fullMetadata,
			}
			allChunks := make([]ContentChunk, 0, len(chunks)+1)
			allChunks = append(allChunks, fullChunk)
			for i, c := range chunks {
				c.Index = i + 1
				allChunks = append(allChunks, c)
			}
			return allChunks
		}
	}

	return chunks
}

func chunkPages(result IngestionResult, config ChunkConfig) []ContentChunk {
	chunks := make([]ContentChunk, 0)
	for _, page := range result.Pages {
		pageChunks := chunkText(page.Text, config)
		pageIndex := page.Index
		for _, chunk := range pageChunks {
			chunk.Index = len(chunks)
			chunk.SourceType = result.SourceType
			chunk.PageIndex = &pageIndex
			chunk.Metadata = copyMetadata(result.Metadata)
			chunks = append(chunks, chunk)
		}
	}
	return chunks
}

func chunkText(text string, config ChunkConfig) []ContentChunk {
	paragraphs := splitParagraphs(text)
	chunks := make([]ContentChunk, 0)
	var current []string
	currentLength := 0

	flush := func() {
		if len(current) == 0 {
			return
		}
		chunks = append(chunks, ContentChunk{
			Index: len(chunks),
			Text:  joinText(current),
		})

		overlap := overlapText(joinText(current), config.OverlapCharacters)
		current = nil
		currentLength = 0
		if overlap != "" {
			current = append(current, overlap)
			currentLength = len(overlap)
		}
	}

	for _, paragraph := range paragraphs {
		if len(paragraph) > config.MaxCharacters {
			if len(current) > 0 {
				flush()
			}
			for _, piece := range splitLongText(paragraph, config) {
				chunks = append(chunks, ContentChunk{Index: len(chunks), Text: piece})
			}
			current = nil
			currentLength = 0
			continue
		}

		addedLength := len(paragraph)
		if currentLength > 0 {
			addedLength += 2
		}
		if currentLength > 0 && currentLength+addedLength > config.MaxCharacters {
			flush()
		}
		current = append(current, paragraph)
		currentLength += addedLength
	}

	if len(current) > 0 {
		chunks = append(chunks, ContentChunk{Index: len(chunks), Text: joinText(current)})
	}

	return chunks
}

func withChunkMetadata(chunks []ContentChunk, sourceType SourceType, metadata map[string]string) []ContentChunk {
	for index := range chunks {
		chunks[index].Index = index
		chunks[index].SourceType = sourceType
		chunks[index].Metadata = copyMetadata(metadata)
	}
	return chunks
}

func splitParagraphs(text string) []string {
	rawParts := strings.Split(cleanStructuredText(text), "\n\n")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		part = cleanStructuredText(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func splitLongText(text string, config ChunkConfig) []string {
	words := strings.Fields(text)
	chunks := make([]string, 0)
	current := make([]string, 0)
	currentLength := 0

	for _, word := range words {
		addedLength := len(word)
		if currentLength > 0 {
			addedLength++
		}
		if currentLength > 0 && currentLength+addedLength > config.MaxCharacters {
			chunks = append(chunks, strings.Join(current, " "))
			overlap := overlapText(strings.Join(current, " "), config.OverlapCharacters)
			current = nil
			currentLength = 0
			if overlap != "" {
				current = append(current, overlap)
				currentLength = len(overlap)
			}
		}
		current = append(current, word)
		currentLength += addedLength
	}

	if len(current) > 0 {
		chunks = append(chunks, strings.Join(current, " "))
	}

	return chunks
}

func overlapText(text string, maxCharacters int) string {
	text = cleanInlineText(text)
	if maxCharacters <= 0 || len(text) <= maxCharacters {
		if len(text) <= maxCharacters {
			return text
		}
		return ""
	}

	start := len(text) - maxCharacters
	for start > 0 && text[start] != ' ' {
		start++
	}
	return strings.TrimSpace(text[start:])
}

// Why this file exists:
// Embeddings work better on smaller pieces than on whole documents or transcripts.
// This chunker creates deterministic chunks now and keeps source references
// such as page numbers and timestamps for later search and citations.
