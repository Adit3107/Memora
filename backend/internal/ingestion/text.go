package ingestion

import "strings"

func normalizeText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func joinText(parts []string) string {
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = normalizeText(part)
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}

	return strings.Join(cleaned, "\n\n")
}

// Why this file exists:
// Different extractors produce text with different spacing and line-break habits.
// These helpers give the ingestion package one consistent way to join and trim text.
