package ingestion

import "strings"

func CleanIngestionResult(result IngestionResult) (IngestionResult, error) {
	cleaned := result
	cleaned.Title = cleanInlineText(result.Title)
	cleaned.Description = cleanStructuredText(result.Description)
	cleaned.Author = cleanInlineText(result.Author)
	cleaned.SourceURL = strings.TrimSpace(result.SourceURL)
	cleaned.RawText = cleanStructuredText(result.RawText)
	cleaned.CleanText = cleanStructuredText(result.CleanText)
	if cleaned.CleanText == "" {
		cleaned.CleanText = cleaned.RawText
	}

	cleaned.Metadata = copyMetadata(result.Metadata)
	cleaned.Transcript = cleanTranscript(result.Transcript)
	cleaned.Pages = cleanPages(result.Pages)

	if cleaned.CleanText == "" && len(cleaned.Transcript) == 0 && len(cleaned.Pages) == 0 {
		return IngestionResult{}, ErrEmptyContent
	}

	return cleaned, nil
}

func cleanInlineText(value string) string {
	return strings.Join(strings.Fields(removeExtractionArtifacts(value)), " ")
}

func cleanStructuredText(value string) string {
	value = removeExtractionArtifacts(value)
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")

	lines := strings.Split(value, "\n")
	cleanedLines := make([]string, 0, len(lines))
	blankCount := 0
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" {
			blankCount++
			if blankCount <= 1 && len(cleanedLines) > 0 {
				cleanedLines = append(cleanedLines, "")
			}
			continue
		}

		blankCount = 0
		cleanedLines = append(cleanedLines, line)
	}

	for len(cleanedLines) > 0 && cleanedLines[len(cleanedLines)-1] == "" {
		cleanedLines = cleanedLines[:len(cleanedLines)-1]
	}

	return strings.TrimSpace(strings.Join(cleanedLines, "\n"))
}

func removeExtractionArtifacts(value string) string {
	replacer := strings.NewReplacer(
		"\x00", "",
		"\uFFFD", "",
		"\f", "\n\n",
		"\u00a0", " ",
	)
	return replacer.Replace(value)
}

func copyMetadata(metadata map[string]string) map[string]string {
	if len(metadata) == 0 {
		return nil
	}

	copied := make(map[string]string, len(metadata))
	for key, value := range metadata {
		key = cleanInlineText(key)
		value = cleanInlineText(value)
		if key != "" && value != "" {
			copied[key] = value
		}
	}
	return copied
}

func cleanTranscript(transcript []TranscriptSegment) []TranscriptSegment {
	if len(transcript) == 0 {
		return nil
	}

	cleaned := make([]TranscriptSegment, 0, len(transcript))
	for _, segment := range transcript {
		text := cleanInlineText(segment.Text)
		if text == "" {
			continue
		}
		cleaned = append(cleaned, TranscriptSegment{
			StartSeconds: segment.StartSeconds,
			EndSeconds:   segment.EndSeconds,
			Text:         text,
		})
	}
	return cleaned
}

func cleanPages(pages []DocumentPage) []DocumentPage {
	if len(pages) == 0 {
		return nil
	}

	cleaned := make([]DocumentPage, 0, len(pages))
	for _, page := range pages {
		text := cleanStructuredText(page.Text)
		if text == "" {
			continue
		}
		cleaned = append(cleaned, DocumentPage{
			Index: page.Index,
			Text:  text,
		})
	}
	return cleaned
}
