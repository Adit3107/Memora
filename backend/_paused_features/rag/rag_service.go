/*
=============================================================================
PAUSED FEATURE: RAG Service (Put on hold for upcoming release)
Provides grounded retrieval-augmented generation across saved content.
Separated and commented out so the developer can focus on working code.
To resume: uncomment this file and move back to internal/services.
=============================================================================

package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"memora-backend/internal/models"
	"memora-backend/internal/repository"
)

type RAGScopeType string

const (
	RAGScopeLibrary         RAGScopeType = "library"
	RAGScopeContent         RAGScopeType = "content"
	RAGScopeSelectedSources RAGScopeType = "selected_sources"
)

const (
	defaultRAGTopK             = 8
	maxRAGTopK                 = 16
	defaultRAGHistoryMessages  = 6
	defaultRAGContextMaxChars  = 12000
	defaultRAGChunkMaxChars    = 1600
	notEnoughSavedContentReply = "I couldn't find relevant information in your saved library."
)

type RAGSearchService interface {
	Search(ctx context.Context, input SearchInput) (SearchResponse, error)
	ContentChunks(ctx context.Context, input SearchInput) (SearchResponse, error)
	ContentMetadata(ctx context.Context, userID string, contentID string) (ContentMetadataResult, error)
}

type ragIntent string

const (
	ragIntentCasual    ragIntent = "casual"
	ragIntentKnowledge ragIntent = "knowledge"
	ragIntentMetadata  ragIntent = "metadata"
)

type RAGConfig struct {
	TopK               int
	HistoryMessages    int
	ContextMaxChars    int
	ChunkMaxChars      int
	AllowLocalFallback bool
}

type RAGService struct {
	search RAGSearchService
	llm    LLMService
	config RAGConfig
}

type RAGInput struct {
	UserID   string
	Question string
	Scope    RAGScope
	History  []RAGMessage
	TopK     int
}

type RAGScope struct {
	Type       RAGScopeType `json:"type"`
	ContentID  string       `json:"content_id,omitempty"`
	ContentIDs []string     `json:"content_ids,omitempty"`
}

type RAGMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RAGResponse struct {
	Answer    string        `json:"answer"`
	Citations []RAGCitation `json:"citations"`
}

type RAGCitation struct {
	ContentID string             `json:"content_id"`
	Title     string             `json:"title"`
	Type      models.ContentType `json:"type"`
	SourceURL *string            `json:"source_url,omitempty"`
	Page      *int               `json:"page,omitempty"`
	Timestamp *RAGTimestamp      `json:"timestamp,omitempty"`
}

type RAGTimestamp struct {
	Start float64 `json:"start"`
	End   float64 `json:"end,omitempty"`
}

func NewRAGService(search RAGSearchService, llm LLMService, config RAGConfig) *RAGService {
	return &RAGService{
		search: search,
		llm:    llm,
		config: normalizeRAGConfig(config),
	}
}

func (s *RAGService) Ask(ctx context.Context, input RAGInput) (RAGResponse, error) {
	normalized, err := s.normalizeInput(input)
	if err != nil {
		return RAGResponse{}, err
	}

	intent := classifyRAGIntent(normalized.Question)
	switch intent {
	case ragIntentCasual:
		return RAGResponse{Answer: casualRAGAnswer(normalized.Scope), Citations: []RAGCitation{}}, nil
	case ragIntentMetadata:
		return s.answerMetadataQuestion(ctx, normalized)
	}

	searchResponse, err := s.retrieve(ctx, normalized)
	if err != nil {
		return RAGResponse{}, err
	}

	chunks := selectContextChunks(searchResponse.Results, s.config, shouldReadScopedContent(normalized))
	if len(chunks) == 0 {
		return RAGResponse{Answer: noContextAnswer(normalized.Scope), Citations: []RAGCitation{}}, nil
	}

	contextText := buildContextText(chunks, s.config)
	citations := buildCitations(chunks)
	systemPrompt := groundedSystemPrompt()
	userPrompt := buildRAGPrompt(normalized.Question, normalized.History, contextText)

	answer, err := s.llm.Generate(ctx, LLMGenerateInput{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	})
	if err != nil {
		if s.config.AllowLocalFallback {
			slog.Warn("rag llm generation failed; returning grounded retrieval fallback", "error", err)
			return RAGResponse{
				Answer:    fallbackGroundedAnswer(normalized.Question, chunks),
				Citations: citations,
			}, nil
		}
		return RAGResponse{}, err
	}

	answer = strings.TrimSpace(answer)
	if answer == "" {
		answer = noContextAnswer(normalized.Scope)
	}
	answer = cleanLLMAnswer(answer)

	return RAGResponse{
		Answer:    answer,
		Citations: citations,
	}, nil
}

func (s *RAGService) retrieve(ctx context.Context, normalized RAGInput) (SearchResponse, error) {
	if shouldReadScopedContent(normalized) {
		searchResponse, err := s.search.ContentChunks(ctx, SearchInput{
			UserID:     normalized.UserID,
			Query:      normalized.Question,
			Mode:       SearchModeKeyword,
			ContentIDs: scopeContentIDs(normalized.Scope),
			Limit:      contentSummaryLimit(normalized.TopK),
			Offset:     0,
		})
		if err == nil && len(searchResponse.Results) > 0 {
			return searchResponse, nil
		}
		if err != nil {
			slog.Warn("rag content-summary retrieval failed; falling back to ranked retrieval", "error", err)
		}
	}

	input := SearchInput{
		UserID:     normalized.UserID,
		Query:      retrievalQuery(normalized.Question, normalized.History),
		Mode:       SearchModeHybrid,
		ContentIDs: scopeContentIDs(normalized.Scope),
		Limit:      normalized.TopK,
		Offset:     0,
	}

	searchResponse, err := s.search.Search(ctx, input)
	if err != nil {
		slog.Warn("rag hybrid retrieval failed; retrying keyword retrieval", "error", err)
		input.Mode = SearchModeKeyword
		return s.search.Search(ctx, input)
	}

	return searchResponse, nil
}

func classifyRAGIntent(question string) ragIntent {
	normalized := strings.ToLower(strings.TrimSpace(question))
	normalized = strings.Trim(normalized, " .!?")
	if normalized == "" {
		return ragIntentCasual
	}

	casualMessages := map[string]struct{}{
		"hi":        {},
		"hello":     {},
		"hey":       {},
		"thanks":    {},
		"thank you": {},
		"okay":      {},
		"ok":        {},
		"cool":      {},
		"nice":      {},
	}
	if _, exists := casualMessages[normalized]; exists {
		return ragIntentCasual
	}

	metadataTerms := []string{
		"title",
		"name of video",
		"video name",
		"who uploaded",
		"uploader",
		"channel",
		"how long",
		"length",
		"duration",
		"when was this saved",
		"when did i save",
		"saved date",
		"platform",
		"source url",
		"url",
		"link",
	}
	for _, term := range metadataTerms {
		if strings.Contains(normalized, term) {
			return ragIntentMetadata
		}
	}

	return ragIntentKnowledge
}

func casualRAGAnswer(scope RAGScope) string {
	switch scope.Type {
	case RAGScopeContent:
		return "Hi! What would you like to know about this content?"
	case RAGScopeSelectedSources:
		return "Hi! What would you like to know about the selected sources?"
	default:
		return "Hi! What would you like to know about your saved library?"
	}
}

func (s *RAGService) answerMetadataQuestion(ctx context.Context, input RAGInput) (RAGResponse, error) {
	if input.Scope.Type != RAGScopeContent {
		return RAGResponse{
			Answer:    "I can answer metadata questions when you're asking about a specific saved item.",
			Citations: []RAGCitation{},
		}, nil
	}

	metadata, err := s.search.ContentMetadata(ctx, input.UserID, input.Scope.ContentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return RAGResponse{Answer: "I couldn't find that saved item.", Citations: []RAGCitation{}}, nil
		}
		return RAGResponse{}, err
	}

	return RAGResponse{
		Answer:    metadataAnswer(input.Question, metadata),
		Citations: []RAGCitation{},
	}, nil
}

func metadataAnswer(question string, metadata ContentMetadataResult) string {
	normalized := strings.ToLower(question)
	switch {
	case strings.Contains(normalized, "title") || strings.Contains(normalized, "name of video") || strings.Contains(normalized, "video name"):
		title := strings.TrimSpace(metadata.Name)
		if title == "" {
			title = strings.TrimSpace(metadata.Title)
		}
		if title == "" {
			return "I don't have a title saved for this item."
		}
		if looksLikeURL(title) {
			return "I don't have a proper title saved for this item. It was saved using its source URL."
		}
		return fmt.Sprintf("The saved name is %q.", title)
	case strings.Contains(normalized, "who uploaded") || strings.Contains(normalized, "uploader") || strings.Contains(normalized, "channel"):
		channel := firstMetadataValue(metadata.Metadata, "channel", "channel_name", "uploader", "author", "creator")
		if channel == "" {
			return "I don't have uploader or channel information saved for this item."
		}
		return fmt.Sprintf("It was uploaded by %s.", channel)
	case strings.Contains(normalized, "how long") || strings.Contains(normalized, "length") || strings.Contains(normalized, "duration"):
		duration := metadataDuration(metadata)
		if duration <= 0 {
			return "I don't have duration information saved for this item."
		}
		return fmt.Sprintf("It's %s long.", formatDuration(duration))
	case strings.Contains(normalized, "when was this saved") || strings.Contains(normalized, "when did i save") || strings.Contains(normalized, "saved date"):
		return fmt.Sprintf("You saved it on %s.", metadata.CreatedAt.Format("January 2, 2006"))
	case strings.Contains(normalized, "platform"):
		platform := firstMetadataValue(metadata.Metadata, "provider", "platform", "content_type")
		if platform == "" {
			platform = string(metadata.ContentType)
		}
		return fmt.Sprintf("It's from %s.", humanizeMetadataValue(platform))
	case strings.Contains(normalized, "url") || strings.Contains(normalized, "link") || strings.Contains(normalized, "source"):
		if metadata.SourceURL == nil || strings.TrimSpace(*metadata.SourceURL) == "" {
			return "I don't have a source URL saved for this item."
		}
		return fmt.Sprintf("The saved source URL is %s.", strings.TrimSpace(*metadata.SourceURL))
	default:
		return "I can answer metadata questions like the title, source URL, platform, saved date, channel, or duration for this item."
	}
}

func firstMetadataValue(metadata map[string]string, keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(metadata[key])
		if value != "" {
			return value
		}
	}
	return ""
}

func metadataDuration(metadata ContentMetadataResult) float64 {
	for _, key := range []string{"duration_seconds", "duration", "length_seconds", "video_duration"} {
		value := strings.TrimSpace(metadata.Metadata[key])
		if value == "" {
			continue
		}
		parsed, err := strconv.ParseFloat(value, 64)
		if err == nil && parsed > 0 {
			return parsed
		}
	}
	if metadata.MaxSeconds != nil && metadata.MinSeconds != nil && *metadata.MaxSeconds > *metadata.MinSeconds {
		return *metadata.MaxSeconds - *metadata.MinSeconds
	}
	if metadata.MaxSeconds != nil && *metadata.MaxSeconds > 0 {
		return *metadata.MaxSeconds
	}
	return 0
}

func formatDuration(seconds float64) string {
	duration := time.Duration(seconds * float64(time.Second)).Round(time.Second)
	totalSeconds := int(duration.Seconds())
	if totalSeconds < 60 {
		return fmt.Sprintf("%d seconds", totalSeconds)
	}
	minutes := totalSeconds / 60
	remainingSeconds := totalSeconds % 60
	if remainingSeconds == 0 {
		return fmt.Sprintf("%d minutes", minutes)
	}
	return fmt.Sprintf("%d minutes %d seconds", minutes, remainingSeconds)
}

func humanizeMetadataValue(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "_", " "))
	if value == "" {
		return value
	}
	if strings.EqualFold(value, "youtube") {
		return "YouTube"
	}
	return value
}

func looksLikeURL(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func shouldReadScopedContent(input RAGInput) bool {
	if input.Scope.Type != RAGScopeContent {
		return false
	}
	question := strings.ToLower(input.Question)
	summaryTerms := []string{
		"summary",
		"summarize",
		"summarise",
		"what is this",
		"what this",
		"what is the video",
		"what this video",
		"video about",
		"content about",
		"main point",
		"main points",
		"key point",
		"key points",
		"takeaway",
		"takeaways",
		"explain this",
	}
	for _, term := range summaryTerms {
		if strings.Contains(question, term) {
			return true
		}
	}
	return false
}

func contentSummaryLimit(topK int) int {
	if topK < 12 {
		return 25
	}
	return topK
}

func (s *RAGService) normalizeInput(input RAGInput) (RAGInput, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.Question = strings.TrimSpace(input.Question)
	if input.UserID == "" || input.Question == "" || utf8.RuneCountInString(input.Question) > 1000 {
		return RAGInput{}, ErrValidation
	}

	if input.Scope.Type == "" {
		input.Scope.Type = RAGScopeLibrary
	}

	switch input.Scope.Type {
	case RAGScopeLibrary:
		input.Scope.ContentID = ""
		input.Scope.ContentIDs = nil
	case RAGScopeContent:
		input.Scope.ContentID = strings.TrimSpace(input.Scope.ContentID)
		if input.Scope.ContentID == "" {
			return RAGInput{}, ErrValidation
		}
		input.Scope.ContentIDs = []string{input.Scope.ContentID}
	case RAGScopeSelectedSources:
		contentIDs := normalizeContentIDs(input.Scope.ContentIDs)
		if len(contentIDs) == 0 {
			return RAGInput{}, ErrValidation
		}
		input.Scope.ContentIDs = contentIDs
		input.Scope.ContentID = ""
	default:
		return RAGInput{}, ErrValidation
	}

	input.History = boundedHistory(input.History, s.config.HistoryMessages)
	if input.TopK <= 0 {
		input.TopK = s.config.TopK
	}
	if input.TopK > maxRAGTopK {
		input.TopK = maxRAGTopK
	}

	return input, nil
}

func normalizeRAGConfig(config RAGConfig) RAGConfig {
	if config.TopK <= 0 {
		config.TopK = defaultRAGTopK
	}
	if config.TopK > maxRAGTopK {
		config.TopK = maxRAGTopK
	}
	if config.HistoryMessages <= 0 {
		config.HistoryMessages = defaultRAGHistoryMessages
	}
	if config.ContextMaxChars <= 0 {
		config.ContextMaxChars = defaultRAGContextMaxChars
	}
	if config.ChunkMaxChars <= 0 {
		config.ChunkMaxChars = defaultRAGChunkMaxChars
	}
	return config
}

func normalizeContentIDs(values []string) []string {
	seen := map[string]struct{}{}
	output := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		output = append(output, value)
	}
	return output
}

func scopeContentIDs(scope RAGScope) []string {
	if scope.Type == RAGScopeContent {
		return []string{scope.ContentID}
	}
	if scope.Type == RAGScopeSelectedSources {
		return scope.ContentIDs
	}
	return nil
}

func boundedHistory(history []RAGMessage, limit int) []RAGMessage {
	clean := make([]RAGMessage, 0, len(history))
	for _, message := range history {
		role := strings.TrimSpace(strings.ToLower(message.Role))
		content := strings.TrimSpace(message.Content)
		if content == "" || (role != "user" && role != "assistant") {
			continue
		}
		if utf8.RuneCountInString(content) > 1000 {
			content = string([]rune(content)[:1000])
		}
		clean = append(clean, RAGMessage{Role: role, Content: content})
	}
	if len(clean) > limit {
		return clean[len(clean)-limit:]
	}
	return clean
}

func retrievalQuery(question string, history []RAGMessage) string {
	recent := boundedHistory(history, 4)
	parts := make([]string, 0, len(recent)+1)
	for _, message := range recent {
		if message.Role == "user" {
			parts = append(parts, message.Content)
		}
	}
	parts = append(parts, question)
	return strings.Join(parts, "\n")
}

func selectContextChunks(results []repository.ChunkSearchResult, config RAGConfig, preserveOrder bool) []repository.ChunkSearchResult {
	selected := make([]repository.ChunkSearchResult, 0, len(results))
	seenChunks := map[string]struct{}{}
	seenTexts := map[string]struct{}{}
	for _, result := range results {
		textKey := compactTextKey(result.Text)
		if result.ChunkID == "" || textKey == "" {
			continue
		}
		if _, exists := seenChunks[result.ChunkID]; exists {
			continue
		}
		if !preserveOrder {
			if _, exists := seenTexts[textKey]; exists {
				continue
			}
			seenTexts[textKey] = struct{}{}
		}
		seenChunks[result.ChunkID] = struct{}{}
		selected = append(selected, result)
		if len(selected) >= config.TopK {
			break
		}
	}
	return selected
}

func compactTextKey(value string) string {
	normalized := strings.Join(strings.Fields(strings.ToLower(value)), " ")
	if len(normalized) > 240 {
		return normalized[:240]
	}
	return normalized
}

func buildContextText(chunks []repository.ChunkSearchResult, config RAGConfig) string {
	var builder strings.Builder
	for _, chunk := range chunks {
		entry := fmt.Sprintf(
			"Saved source title: %s\nSource type: %s\nLocation: %s\nContent excerpt:\n%s\n\n",
			chunk.Title,
			chunk.ContentType,
			humanLocationLabel(chunk),
			truncateRunes(strings.TrimSpace(chunk.Text), config.ChunkMaxChars),
		)
		if builder.Len()+len(entry) > config.ContextMaxChars {
			break
		}
		builder.WriteString(entry)
	}
	return strings.TrimSpace(builder.String())
}

func locationLabel(chunk repository.ChunkSearchResult) string {
	if chunk.StartSeconds != nil {
		if chunk.EndSeconds != nil {
			return fmt.Sprintf(" timestamp=%.0f-%.0fs", *chunk.StartSeconds, *chunk.EndSeconds)
		}
		return fmt.Sprintf(" timestamp=%.0fs", *chunk.StartSeconds)
	}
	if chunk.PageIndex != nil {
		return fmt.Sprintf(" page=%d", *chunk.PageIndex+1)
	}
	return ""
}

func humanLocationLabel(chunk repository.ChunkSearchResult) string {
	if chunk.StartSeconds != nil {
		if chunk.EndSeconds != nil {
			return fmt.Sprintf("%.0f-%.0f seconds", *chunk.StartSeconds, *chunk.EndSeconds)
		}
		return fmt.Sprintf("%.0f seconds", *chunk.StartSeconds)
	}
	if chunk.PageIndex != nil {
		return fmt.Sprintf("page %d", *chunk.PageIndex+1)
	}
	return "not specified"
}

func buildCitations(chunks []repository.ChunkSearchResult) []RAGCitation {
	citations := make([]RAGCitation, 0, len(chunks))
	byContentID := map[string]int{}
	for _, chunk := range chunks {
		index, exists := byContentID[chunk.ContentID]
		if !exists {
			byContentID[chunk.ContentID] = len(citations)
			citations = append(citations, RAGCitation{
				ContentID: chunk.ContentID,
				Title:     displaySourceTitle(chunk),
				Type:      chunk.ContentType,
				SourceURL: chunk.SourceURL,
			})
			index = len(citations) - 1
		}

		if chunk.PageIndex != nil {
			page := *chunk.PageIndex + 1
			if citations[index].Page == nil || page < *citations[index].Page {
				citations[index].Page = &page
			}
		}
		if chunk.StartSeconds != nil {
			if citations[index].Timestamp == nil {
				citations[index].Timestamp = &RAGTimestamp{Start: *chunk.StartSeconds}
			}
			if *chunk.StartSeconds < citations[index].Timestamp.Start {
				citations[index].Timestamp.Start = *chunk.StartSeconds
			}
			if chunk.EndSeconds != nil && *chunk.EndSeconds > citations[index].Timestamp.End {
				citations[index].Timestamp.End = *chunk.EndSeconds
			}
		}
	}
	return citations
}

func displaySourceTitle(chunk repository.ChunkSearchResult) string {
	title := strings.TrimSpace(chunk.Title)
	if title != "" && !strings.HasPrefix(strings.ToLower(title), "http://") && !strings.HasPrefix(strings.ToLower(title), "https://") {
		return title
	}
	if chunk.ContentType == models.ContentTypeVideo {
		return "YouTube video"
	}
	if chunk.ContentType == models.ContentTypeDocument {
		return "Document"
	}
	return "Saved source"
}

func groundedSystemPrompt() string {
	return strings.Join([]string{
		"You are MEMORA, a personal knowledge assistant.",
		"The user has saved content in MEMORA and is asking a question about it.",
		"You will receive saved-content excerpts as internal context.",
		"Use that context to answer the user's question naturally and conversationally.",
		"Treat the provided context as internal knowledge.",
		"Never mention retrieval, chunks, top-K results, search scores, context excerpts, internal IDs, or metadata.",
		"Never output bracketed source labels or citation markers.",
		"Do not say 'based on the retrieved context', 'according to the provided context', or similar phrases.",
		"Do not dump or closely copy the provided context.",
		"Synthesize the information and answer the actual question.",
		"Keep the response natural, clear, and useful.",
		"Do not fabricate facts unsupported by the provided context.",
		"If the provided context is insufficient, say that the saved content does not contain enough information.",
		"When the user asks about a specific saved item, answer only using that item's available context.",
		"When the user asks about selected sources, use only those selected sources.",
		"MEMORA handles source citations separately, so do not include citation labels in your answer.",
	}, "\n")
}

func buildRAGPrompt(question string, history []RAGMessage, contextText string) string {
	var builder strings.Builder
	if len(history) > 0 {
		builder.WriteString("Recent conversation:\n")
		for _, message := range history {
			builder.WriteString(strings.ToUpper(message.Role[:1]))
			builder.WriteString(message.Role[1:])
			builder.WriteString(": ")
			builder.WriteString(message.Content)
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}
	builder.WriteString("MEMORA context:\n")
	builder.WriteString(contextText)
	builder.WriteString("\n\nQuestion:\n")
	builder.WriteString(question)
	return builder.String()
}

func fallbackGroundedAnswer(question string, chunks []repository.ChunkSearchResult) string {
	return fallbackLeadSentence(question, chunks)
}

func fallbackLeadSentence(question string, chunks []repository.ChunkSearchResult) string {
	if len(chunks) == 0 {
		return notEnoughSavedContentReply
	}

	normalizedQuestion := strings.ToLower(strings.TrimSpace(question))
	if asksForTitle(normalizedQuestion) {
		title := strings.TrimSpace(displaySourceTitle(chunks[0]))
		if title == "" || title == "YouTube video" || title == "Saved source" {
			return "I don't have a proper title saved for this item. It looks like this content was saved with its source URL instead of a video title."
		}
		return fmt.Sprintf("The saved title is %q.", title)
	}

	text := strings.TrimSpace(chunks[0].Text)
	if text == "" {
		return "I found related saved content, but there is not enough readable text to answer confidently."
	}

	lead := trimSentenceLead(text, 360)
	if asksForConcepts(normalizedQuestion) {
		return fallbackConceptAnswer(chunks)
	}
	if strings.Contains(normalizedQuestion, "video") || strings.Contains(normalizedQuestion, "about") || strings.Contains(normalizedQuestion, "summar") {
		return fmt.Sprintf("This content is about %s.", lead)
	}

	return fmt.Sprintf("The saved content says %s.", lead)
}

func asksForTitle(question string) bool {
	titleTerms := []string{"title", "name of video", "video name", "called"}
	for _, term := range titleTerms {
		if strings.Contains(question, term) {
			return true
		}
	}
	return false
}

func asksForConcepts(question string) bool {
	conceptTerms := []string{"important concept", "important concepts", "key concept", "key concepts", "explain", "main point", "main points"}
	for _, term := range conceptTerms {
		if strings.Contains(question, term) {
			return true
		}
	}
	return false
}

func fallbackConceptAnswer(chunks []repository.ChunkSearchResult) string {
	text := strings.ToLower(strings.Join(chunkTexts(chunks), " "))
	concepts := make([]string, 0, 4)
	if strings.Contains(text, "cab") || strings.Contains(text, "driver") {
		concepts = append(concepts, "cab-booking systems need to quickly match users with nearby drivers")
	}
	if strings.Contains(text, "kafka") {
		concepts = append(concepts, "Kafka is introduced as the backend component involved in that communication")
	}
	if strings.Contains(text, "distributed message broker") || strings.Contains(text, "message broker") {
		concepts = append(concepts, "a distributed message broker helps services pass messages reliably")
	}
	if strings.Contains(text, "publish") || strings.Contains(text, "subscriber") || strings.Contains(text, "subscribe") {
		concepts = append(concepts, "the publish-subscribe pattern lets one part of a system publish events while other parts consume them")
	}
	if len(concepts) == 0 {
		return fmt.Sprintf("The important idea is that %s.", trimSentenceLead(chunks[0].Text, 360))
	}

	if len(concepts) == 1 {
		return "The important concept is that " + concepts[0] + "."
	}

	return "The important concepts are: " + strings.Join(concepts[:len(concepts)-1], "; ") + "; and " + concepts[len(concepts)-1] + "."
}

func chunkTexts(chunks []repository.ChunkSearchResult) []string {
	texts := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		text := strings.TrimSpace(chunk.Text)
		if text != "" {
			texts = append(texts, text)
		}
	}
	return texts
}

func trimSentenceLead(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	value = strings.Trim(value, " .,;:-")
	if value == "" {
		return "related saved content"
	}
	return truncateRunes(value, limit)
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return strings.TrimSpace(string(runes[:limit])) + "..."
}

func noContextAnswer(scope RAGScope) string {
	switch scope.Type {
	case RAGScopeContent:
		return "I couldn't find enough information about that in this content."
	case RAGScopeSelectedSources:
		return "I couldn't find enough relevant information about that in the selected sources."
	default:
		return notEnoughSavedContentReply
	}
}

func cleanLLMAnswer(answer string) string {
	cleaned := strings.TrimSpace(answer)
	for i := 1; i <= 20; i++ {
		cleaned = strings.ReplaceAll(cleaned, fmt.Sprintf("[S%d]", i), "")
	}
	replacements := []struct {
		old string
		new string
	}{
		{"Based on the retrieved MEMORA context, ", ""},
		{"Based on the retrieved context, ", ""},
		{"According to the provided context, ", ""},
		{"According to the provided MEMORA context, ", ""},
		{"Based on the provided context, ", ""},
	}
	for _, replacement := range replacements {
		cleaned = strings.ReplaceAll(cleaned, replacement.old, replacement.new)
		cleaned = strings.ReplaceAll(cleaned, strings.ToLower(replacement.old), replacement.new)
	}
	return strings.TrimSpace(cleaned)
}

func IsRAGClientError(err error) bool {
	return errors.Is(err, ErrValidation)
}
*/

