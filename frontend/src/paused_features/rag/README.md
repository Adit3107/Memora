# Paused Feature: AI Playground & RAG Chat UI

> **Status:** PAUSED / ON HOLD  
> **Target Release:** Upcoming Release

This directory isolates the AI Playground and RAG Chat components while the backend RAG service is paused.

## Contained Files

- **`ai-playground.tsx`**: Full interactive AI Playground featuring:
  - Library-wide search & synthesis
  - Single-item Q&A
  - Selected multi-source comparison
  - Grounded citations with YouTube timestamp deep-links and document page markers
- **`rag-chat-panel.tsx`**: Scoped contextual chat panel for space and document views.

## How to Resume & Re-enable

1. Move `ai-playground.tsx` and `rag-chat-panel.tsx` back to `src/components/rag/`.
2. Remove the outer `/* ... */` comment blocks in both files.
3. In `src/lib/api.ts`, uncomment the RAG TypeScript types and `askMemora()` function.
4. In `src/app/app/ai/page.tsx`, re-import and render `<AIPlayground />`.
