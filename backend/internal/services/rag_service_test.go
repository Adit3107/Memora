package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"memora-backend/internal/models"
	"memora-backend/internal/repository"
)

type fakeRAGSearch struct {
	lastInput      SearchInput
	calls          []SearchInput
	contentCalls   []SearchInput
	results        []repository.ChunkSearchResult
	contentResults []repository.ChunkSearchResult
	metadata       repository.ContentMetadataResult
	err            error
	contentErr     error
	metadataErr    error
	errs           []error
}

func (f *fakeRAGSearch) Search(_ context.Context, input SearchInput) (SearchResponse, error) {
	f.lastInput = input
	f.calls = append(f.calls, input)
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		if err != nil {
			return SearchResponse{}, err
		}
	}
	if f.err != nil {
		return SearchResponse{}, f.err
	}
	return SearchResponse{Mode: input.Mode, Query: input.Query, Total: len(f.results), Results: f.results}, nil
}

func (f *fakeRAGSearch) ContentChunks(_ context.Context, input SearchInput) (SearchResponse, error) {
	f.lastInput = input
	f.contentCalls = append(f.contentCalls, input)
	if f.contentErr != nil {
		return SearchResponse{}, f.contentErr
	}
	results := f.contentResults
	if results == nil {
		results = f.results
	}
	return SearchResponse{Mode: input.Mode, Query: input.Query, Total: len(results), Results: results}, nil
}

func (f *fakeRAGSearch) ContentMetadata(_ context.Context, userID string, contentID string) (ContentMetadataResult, error) {
	if f.metadataErr != nil {
		return ContentMetadataResult{}, f.metadataErr
	}
	metadata := f.metadata
	if metadata.ContentID == "" {
		metadata = repository.ContentMetadataResult{
			ContentID:   contentID,
			UserID:      userID,
			Title:       "Kafka Short",
			ContentType: models.ContentTypeVideo,
			CreatedAt:   time.Date(2026, time.September, 23, 0, 0, 0, 0, time.UTC),
			Metadata:    map[string]string{"provider": "youtube"},
		}
	}
	return metadata, nil
}

type fakeLLM struct {
	answer string
	err    error
	input  LLMGenerateInput
}

func TestRAGCasualQuestionDoesNotRetrieve(t *testing.T) {
	search := &fakeRAGSearch{results: []repository.ChunkSearchResult{videoChunk("content-a", "chunk-a")}}
	service := NewRAGService(search, &fakeLLM{answer: "should not run"}, RAGConfig{})

	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "hi",
		Scope:    RAGScope{Type: RAGScopeContent, ContentID: "content-a"},
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if got.Answer != "Hi! What would you like to know about this content?" {
		t.Fatalf("unexpected casual answer: %q", got.Answer)
	}
	if len(search.calls) != 0 || len(search.contentCalls) != 0 {
		t.Fatalf("casual question should not retrieve, search calls=%d content calls=%d", len(search.calls), len(search.contentCalls))
	}
}

func (f *fakeLLM) Generate(_ context.Context, input LLMGenerateInput) (string, error) {
	f.input = input
	if f.err != nil {
		return "", f.err
	}
	return f.answer, nil
}

func TestRAGLibraryScopeUsesAuthenticatedUserAndNoContentFilter(t *testing.T) {
	search := &fakeRAGSearch{results: []repository.ChunkSearchResult{videoChunk("content-a", "chunk-a")}}
	llm := &fakeLLM{answer: "You saved content about Kafka's core messaging concepts."}
	service := NewRAGService(search, llm, RAGConfig{})

	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "What did I save about Kafka?",
		Scope:    RAGScope{Type: RAGScopeLibrary},
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}

	if search.lastInput.UserID != "user-a" {
		t.Fatalf("expected user-a retrieval, got %q", search.lastInput.UserID)
	}
	if len(search.lastInput.ContentIDs) != 0 {
		t.Fatalf("expected no content filter for library scope, got %#v", search.lastInput.ContentIDs)
	}
	if got.Answer != "You saved content about Kafka's core messaging concepts." || len(got.Citations) != 1 {
		t.Fatalf("unexpected response: %#v", got)
	}
}

func TestRAGContentScopeRestrictsRetrievalToOneContentID(t *testing.T) {
	search := &fakeRAGSearch{results: []repository.ChunkSearchResult{videoChunk("content-x", "chunk-x")}}
	service := NewRAGService(search, &fakeLLM{answer: "This video introduces Kafka as a messaging system."}, RAGConfig{})

	_, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "Summarize this",
		Scope:    RAGScope{Type: RAGScopeContent, ContentID: " content-x "},
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}

	if len(search.lastInput.ContentIDs) != 1 || search.lastInput.ContentIDs[0] != "content-x" {
		t.Fatalf("expected only content-x, got %#v", search.lastInput.ContentIDs)
	}
}

func TestRAGContentSummaryReadsOrderedChunksFromCurrentContent(t *testing.T) {
	search := &fakeRAGSearch{
		contentResults: []repository.ChunkSearchResult{
			videoChunkWithTimes("content-x", "chunk-1", 0, 20),
			videoChunkWithTimes("content-x", "chunk-2", 20, 51),
		},
		results: []repository.ChunkSearchResult{videoChunk("other-content", "other-chunk")},
	}
	service := NewRAGService(search, &fakeLLM{answer: "This video explains Kafka with a cab-booking example."}, RAGConfig{})

	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "What is this video about?",
		Scope:    RAGScope{Type: RAGScopeContent, ContentID: "content-x"},
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if len(search.contentCalls) != 1 {
		t.Fatalf("expected content chunk retrieval, got %d calls", len(search.contentCalls))
	}
	if len(search.calls) != 0 {
		t.Fatalf("expected no ranked search when content chunks are available, got %d calls", len(search.calls))
	}
	if len(search.contentCalls[0].ContentIDs) != 1 || search.contentCalls[0].ContentIDs[0] != "content-x" {
		t.Fatalf("expected only content-x, got %#v", search.contentCalls[0].ContentIDs)
	}
	if got.Answer != "This video explains Kafka with a cab-booking example." {
		t.Fatalf("unexpected answer: %q", got.Answer)
	}
	if len(got.Citations) != 1 || got.Citations[0].Timestamp == nil || got.Citations[0].Timestamp.Start != 0 || got.Citations[0].Timestamp.End != 51 {
		t.Fatalf("expected merged full-video timestamp citation, got %#v", got.Citations)
	}
}

func TestRAGMetadataTitleBypassesRetrieval(t *testing.T) {
	search := &fakeRAGSearch{
		metadata: repository.ContentMetadataResult{
			ContentID:   "content-a",
			UserID:      "user-a",
			Name:        "Kafka Cab Booking Explained",
			Title:       "Kafka Cab Booking Explained",
			ContentType: models.ContentTypeVideo,
			Metadata:    map[string]string{"provider": "youtube"},
		},
		results: []repository.ChunkSearchResult{videoChunk("content-a", "chunk-a")},
	}
	service := NewRAGService(search, &fakeLLM{answer: "should not run"}, RAGConfig{})

	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "What's the title?",
		Scope:    RAGScope{Type: RAGScopeContent, ContentID: "content-a"},
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if got.Answer != "The saved name is \"Kafka Cab Booking Explained\"." {
		t.Fatalf("unexpected title answer: %q", got.Answer)
	}
	if len(search.calls) != 0 || len(search.contentCalls) != 0 {
		t.Fatalf("metadata question should bypass retrieval, search calls=%d content calls=%d", len(search.calls), len(search.contentCalls))
	}
}

func TestRAGMetadataDurationUsesChunkTimestamps(t *testing.T) {
	minSeconds := 0.0
	maxSeconds := 51.0
	search := &fakeRAGSearch{
		metadata: repository.ContentMetadataResult{
			ContentID:   "content-a",
			UserID:      "user-a",
			Title:       "Kafka Short",
			ContentType: models.ContentTypeVideo,
			MinSeconds:  &minSeconds,
			MaxSeconds:  &maxSeconds,
			Metadata:    map[string]string{"provider": "youtube"},
		},
	}
	service := NewRAGService(search, &fakeLLM{answer: "should not run"}, RAGConfig{})

	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "What's the length of the video?",
		Scope:    RAGScope{Type: RAGScopeContent, ContentID: "content-a"},
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if got.Answer != "It's 51 seconds long." {
		t.Fatalf("unexpected duration answer: %q", got.Answer)
	}
	if len(search.calls) != 0 || len(search.contentCalls) != 0 {
		t.Fatalf("duration question should bypass retrieval, search calls=%d content calls=%d", len(search.calls), len(search.contentCalls))
	}
}

func TestRAGSelectedSourcesDeduplicatesAndRestrictsContentIDs(t *testing.T) {
	search := &fakeRAGSearch{results: []repository.ChunkSearchResult{documentChunk("content-y", "chunk-y")}}
	service := NewRAGService(search, &fakeLLM{answer: "The selected sources focus on Kafka consumer behavior."}, RAGConfig{})

	_, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "Compare these",
		Scope: RAGScope{
			Type:       RAGScopeSelectedSources,
			ContentIDs: []string{"content-x", "content-y", "content-x", " "},
		},
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}

	want := []string{"content-x", "content-y"}
	if strings.Join(search.lastInput.ContentIDs, ",") != strings.Join(want, ",") {
		t.Fatalf("expected %v, got %v", want, search.lastInput.ContentIDs)
	}
}

func TestRAGSelectedSourcesRejectsEmptyScope(t *testing.T) {
	service := NewRAGService(&fakeRAGSearch{}, &fakeLLM{}, RAGConfig{})
	_, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "Compare",
		Scope:    RAGScope{Type: RAGScopeSelectedSources},
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestRAGNoResultsReturnsGroundedNotFound(t *testing.T) {
	service := NewRAGService(&fakeRAGSearch{}, &fakeLLM{answer: "should not be used"}, RAGConfig{})
	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "Unknown topic",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if got.Answer != notEnoughSavedContentReply || len(got.Citations) != 0 {
		t.Fatalf("unexpected no-result response: %#v", got)
	}
}

func TestRAGCitationsIncludeVideoTimestampAndDocumentPage(t *testing.T) {
	search := &fakeRAGSearch{results: []repository.ChunkSearchResult{
		videoChunk("video-1", "chunk-video"),
		documentChunk("doc-1", "chunk-doc"),
	}}
	service := NewRAGService(search, &fakeLLM{answer: "The saved content covers Kafka and consumer groups."}, RAGConfig{})
	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "What is here?",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if got.Citations[0].Timestamp == nil || got.Citations[0].Timestamp.Start != 272 {
		t.Fatalf("expected video timestamp citation, got %#v", got.Citations[0])
	}
	if got.Citations[1].Page == nil || *got.Citations[1].Page != 12 {
		t.Fatalf("expected document page citation, got %#v", got.Citations[1])
	}
}

func TestRAGFollowUpUsesHistoryForFreshRetrieval(t *testing.T) {
	search := &fakeRAGSearch{results: []repository.ChunkSearchResult{videoChunk("content-a", "chunk-a")}}
	service := NewRAGService(search, &fakeLLM{answer: "Kafka's main components include topics, partitions, producers, and consumers."}, RAGConfig{})

	_, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "What are its main components?",
		History: []RAGMessage{
			{Role: "user", Content: "What is Kafka?"},
			{Role: "assistant", Content: "Kafka is an event streaming platform."},
		},
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if !strings.Contains(search.lastInput.Query, "What is Kafka?") || !strings.Contains(search.lastInput.Query, "What are its main components?") {
		t.Fatalf("expected history and question in retrieval query, got %q", search.lastInput.Query)
	}
}

func TestRAGLocalFallbackUsesRetrievedContextWhenLLMUnavailable(t *testing.T) {
	search := &fakeRAGSearch{results: []repository.ChunkSearchResult{videoChunk("content-a", "chunk-a")}}
	service := NewRAGService(search, &fakeLLM{err: ErrLLMUnavailable}, RAGConfig{AllowLocalFallback: true})
	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "Summarize Kafka",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	assertNaturalAnswer(t, got.Answer)
	if !strings.Contains(got.Answer, "This content is about") || len(got.Citations) != 1 {
		t.Fatalf("unexpected fallback response: %#v", got)
	}
}

func TestRAGLLMProviderFailureUsesLocalFallbackWhenEnabled(t *testing.T) {
	search := &fakeRAGSearch{results: []repository.ChunkSearchResult{videoChunk("content-a", "chunk-a")}}
	service := NewRAGService(search, &fakeLLM{err: errors.New("provider returned 403")}, RAGConfig{AllowLocalFallback: true})
	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "Summarize Kafka",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	assertNaturalAnswer(t, got.Answer)
	if !strings.Contains(got.Answer, "This content is about") || len(got.Citations) != 1 {
		t.Fatalf("unexpected fallback response: %#v", got)
	}
}

func TestRAGFallbackTitleQuestionDoesNotReturnTranscriptSummary(t *testing.T) {
	search := &fakeRAGSearch{results: []repository.ChunkSearchResult{videoChunk("content-a", "chunk-a")}}
	service := NewRAGService(search, &fakeLLM{err: ErrLLMUnavailable}, RAGConfig{AllowLocalFallback: true})
	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "whats the title of video",
		Scope:    RAGScope{Type: RAGScopeContent, ContentID: "content-a"},
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	assertNaturalAnswer(t, got.Answer)
	if strings.Contains(strings.ToLower(got.Answer), "cab") || strings.Contains(strings.ToLower(got.Answer), "kafka uses topics") {
		t.Fatalf("title fallback should not summarize transcript: %q", got.Answer)
	}
}

func TestRAGFallbackConceptQuestionSynthesizesInsteadOfRepeatingTranscript(t *testing.T) {
	search := &fakeRAGSearch{results: []repository.ChunkSearchResult{
		{
			ContentID:    "content-a",
			ChunkID:      "chunk-a",
			Title:        "https://youtube.com/shorts/example",
			ContentType:  models.ContentTypeVideo,
			Text:         "when booking any cab service they assign a nearby driver with Kafka. Kafka is a distributed message broker or publisher subscriber system.",
			StartSeconds: floatPtr(0),
			EndSeconds:   floatPtr(51),
		},
	}}
	service := NewRAGService(search, &fakeLLM{err: ErrLLMUnavailable}, RAGConfig{AllowLocalFallback: true})
	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "Explain the important concepts.",
		Scope:    RAGScope{Type: RAGScopeContent, ContentID: "content-a"},
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	assertNaturalAnswer(t, got.Answer)
	if !strings.Contains(got.Answer, "cab-booking systems") || !strings.Contains(got.Answer, "publish-subscribe") {
		t.Fatalf("expected synthesized concept fallback, got %q", got.Answer)
	}
	if strings.Contains(got.Answer, "when booking any cab service") {
		t.Fatalf("concept fallback repeated transcript: %q", got.Answer)
	}
}

func TestRAGRetriesKeywordRetrievalWhenHybridFails(t *testing.T) {
	search := &fakeRAGSearch{
		errs:    []error{errors.New("embedding service unavailable"), nil},
		results: []repository.ChunkSearchResult{documentChunk("doc-1", "chunk-doc")},
	}
	service := NewRAGService(search, &fakeLLM{answer: "Consumer groups divide partition work across consumers."}, RAGConfig{})
	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "consumer groups",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if len(search.calls) != 2 {
		t.Fatalf("expected hybrid and keyword retrieval attempts, got %d", len(search.calls))
	}
	if search.calls[0].Mode != SearchModeHybrid || search.calls[1].Mode != SearchModeKeyword {
		t.Fatalf("unexpected retrieval modes: %#v", search.calls)
	}
	if got.Answer != "Consumer groups divide partition work across consumers." {
		t.Fatalf("unexpected answer: %#v", got)
	}
}

func TestRAGCleansCitationLabelsFromLLMAnswer(t *testing.T) {
	search := &fakeRAGSearch{results: []repository.ChunkSearchResult{videoChunk("content-a", "chunk-a")}}
	service := NewRAGService(search, &fakeLLM{answer: "Kafka uses topics and producers. [S1]"}, RAGConfig{})
	got, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "What is this about?",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	assertNaturalAnswer(t, got.Answer)
	if strings.Contains(got.Answer, "[S1]") {
		t.Fatalf("answer leaked citation marker: %q", got.Answer)
	}
}

func TestRAGPromptDoesNotAskLLMForInlineCitations(t *testing.T) {
	search := &fakeRAGSearch{results: []repository.ChunkSearchResult{videoChunk("content-a", "chunk-a")}}
	llm := &fakeLLM{answer: "This video explains Kafka messaging."}
	service := NewRAGService(search, llm, RAGConfig{})
	_, err := service.Ask(context.Background(), RAGInput{
		UserID:   "user-a",
		Question: "What is this video about?",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if strings.Contains(llm.input.SystemPrompt, "[S1]") || strings.Contains(llm.input.UserPrompt, "[S1]") {
		t.Fatalf("prompt should not contain bracket citation labels:\nSYSTEM:\n%s\nUSER:\n%s", llm.input.SystemPrompt, llm.input.UserPrompt)
	}
	if strings.Contains(strings.ToLower(llm.input.SystemPrompt), "cite relevant sources") {
		t.Fatalf("prompt should not ask LLM to generate citations: %s", llm.input.SystemPrompt)
	}
}

func assertNaturalAnswer(t *testing.T, answer string) {
	t.Helper()
	forbidden := []string{
		"Based on the retrieved",
		"retrieved MEMORA context",
		"retrieved context",
		"Relevant saved notes",
		"Retrieved chunks",
		"Search results",
		"[S1]",
		"[S2]",
		"LLM provider",
	}
	for _, phrase := range forbidden {
		if strings.Contains(answer, phrase) {
			t.Fatalf("answer contains forbidden phrase %q: %q", phrase, answer)
		}
	}
}

func videoChunk(contentID string, chunkID string) repository.ChunkSearchResult {
	return videoChunkWithTimes(contentID, chunkID, 272, 315)
}

func videoChunkWithTimes(contentID string, chunkID string, startValue float64, endValue float64) repository.ChunkSearchResult {
	start := startValue
	end := endValue
	return repository.ChunkSearchResult{
		ContentID:    contentID,
		ChunkID:      chunkID,
		ChunkIndex:   3,
		Title:        "Kafka Introduction",
		ContentType:  models.ContentTypeVideo,
		Text:         "Kafka uses topics, partitions, producers, and consumers.",
		Score:        0.91,
		StartSeconds: &start,
		EndSeconds:   &end,
		SourceType:   "youtube",
		Metadata:     map[string]string{},
	}
}

func documentChunk(contentID string, chunkID string) repository.ChunkSearchResult {
	pageIndex := 11
	return repository.ChunkSearchResult{
		ContentID:   contentID,
		ChunkID:     chunkID,
		ChunkIndex:  7,
		Title:       "Kafka Notes.pdf",
		ContentType: models.ContentTypeDocument,
		Text:        "Consumer groups divide partition work across consumers.",
		Score:       0.82,
		PageIndex:   &pageIndex,
		SourceType:  "document",
		Metadata:    map[string]string{},
	}
}

func floatPtr(value float64) *float64 {
	return &value
}
