/*
=============================================================================
PAUSED FEATURE: RAG Chat Panel (Put on hold for upcoming release)
Reusable RAG chat panel for asking questions scoped to content or spaces.
Separated and commented out so the developer can focus on working code.
To resume: uncomment this file and re-mount where needed.
=============================================================================

"use client";

import { Bot, FileText, Library, Send, Video } from "lucide-react";
import { FormEvent, useMemo, useState } from "react";

import {
	askMemora,
	MEMORA_DEMO_USER_ID,
	type RAGCitation,
	type RAGMessage,
	type RAGScope,
} from "@/lib/api";

type RAGChatPanelProps = {
	title: string;
	scope: RAGScope;
	placeholder?: string;
	disabled?: boolean;
	disabledReason?: string;
	suggestions?: string[];
};

type ThreadMessage = RAGMessage & {
	citations?: RAGCitation[];
};

const defaultSuggestions = [
	"What are the key points?",
	"Summarize this clearly.",
	"What should I remember?",
];

export function RAGChatPanel({
	title,
	scope,
	placeholder = "Ask your saved knowledge...",
	disabled = false,
	disabledReason,
	suggestions = defaultSuggestions,
}: RAGChatPanelProps) {
	const [question, setQuestion] = useState("");
	const [messages, setMessages] = useState<ThreadMessage[]>([]);
	const [isAsking, setIsAsking] = useState(false);
	const [error, setError] = useState("");

	const history = useMemo(
		() =>
			messages
				.slice(-6)
				.map((message) => ({ role: message.role, content: message.content })),
		[messages]
	);

	async function submitQuestion(event?: FormEvent<HTMLFormElement>, value = question) {
		event?.preventDefault();
		const trimmedQuestion = value.trim();
		if (!trimmedQuestion || disabled) {
			return;
		}

		setQuestion("");
		setError("");
		setIsAsking(true);
		setMessages((current) => [
			...current,
			{ role: "user", content: trimmedQuestion },
		]);

		try {
			const result = await askMemora({
				user_id: MEMORA_DEMO_USER_ID,
				question: trimmedQuestion,
				scope,
				history,
			});
			setMessages((current) => [
				...current,
				{
					role: "assistant",
					content: result.answer,
					citations: result.citations,
				},
			]);
		} catch {
			setError("MEMORA could not answer that question. Check the backend, AI service, and LLM configuration.");
		} finally {
			setIsAsking(false);
		}
	}

	return (
		<section className="rounded-md border bg-card p-5 shadow-sm">
			<div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
				<div className="flex items-center gap-3">
					<span className="flex size-9 items-center justify-center rounded-md border bg-background">
						<Bot className="size-4 text-primary" aria-hidden="true" />
					</span>
					<div>
						<h2 className="text-lg font-semibold">{title}</h2>
						{disabled && disabledReason ? (
							<p className="text-sm text-muted-foreground">{disabledReason}</p>
						) : null}
					</div>
				</div>
			</div>

			{messages.length > 0 ? (
				<div className="mt-4 space-y-4">
					{messages.map((message, index) => (
						<div
							className={[
								"rounded-md border p-4",
								message.role === "user" ? "bg-background" : "bg-secondary/50",
							].join(" ")}
							key={`${message.role}-${index}`}
						>
							<p className="text-xs font-medium uppercase text-muted-foreground">
								{message.role === "user" ? "You" : "MEMORA"}
							</p>
							<p className="mt-2 whitespace-pre-wrap text-sm leading-6">
								{message.content}
							</p>
							{message.citations && message.citations.length > 0 ? (
								<CitationList citations={message.citations} />
							) : null}
						</div>
					))}
				</div>
			) : (
				<div className="mt-4 flex flex-wrap gap-2">
					{suggestions.map((suggestion) => (
						<button
							className="rounded-md border bg-background px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground disabled:cursor-not-allowed disabled:opacity-50"
							disabled={disabled || isAsking}
							key={suggestion}
							onClick={() => void submitQuestion(undefined, suggestion)}
							type="button"
						>
							{suggestion}
						</button>
					))}
				</div>
			)}

			{error ? <p className="mt-4 text-sm text-destructive">{error}</p> : null}

			<form className="mt-4 flex flex-col gap-3 sm:flex-row" onSubmit={submitQuestion}>
				<div className="flex min-h-11 flex-1 items-center gap-2 rounded-md border bg-background px-3">
					<Library className="size-4 text-muted-foreground" aria-hidden="true" />
					<input
						className="h-11 w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
						disabled={disabled || isAsking}
						onChange={(event) => setQuestion(event.target.value)}
						placeholder={placeholder}
						value={question}
					/>
				</div>
				<button
					className="inline-flex h-11 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
					disabled={disabled || isAsking || !question.trim()}
					type="submit"
				>
					<Send className="size-4" aria-hidden="true" />
					{isAsking ? "Asking" : "Ask"}
				</button>
			</form>
		</section>
	);
}

function CitationList({ citations }: { citations: RAGCitation[] }) {
	return (
		<div className="mt-4 border-t pt-3">
			<p className="text-xs font-medium uppercase text-muted-foreground">Sources</p>
			<div className="mt-2 grid gap-2">
				{citations.map((citation, index) => {
					const Icon = citation.type === "video" ? Video : FileText;
					const label = citation.timestamp
						? formatTimestampRange(citation.timestamp.start, citation.timestamp.end)
						: citation.page
							? `Page ${citation.page}`
							: "Saved source";
					const href = citationHref(citation);

					return (
						<a
							className="flex items-start gap-3 rounded-md border bg-background p-3 text-sm transition-colors hover:bg-accent"
							href={href}
							key={`${citation.content_id}-${index}`}
							rel="noreferrer"
							target={citation.source_url ? "_blank" : undefined}
						>
							<Icon className="mt-0.5 size-4 text-primary" aria-hidden="true" />
							<span className="min-w-0">
								<span className="block truncate font-medium">{citation.title}</span>
								<span className="block text-xs text-muted-foreground">{label}</span>
							</span>
						</a>
					);
				})}
			</div>
		</div>
	);
}

function formatTimestamp(value: number) {
	const totalSeconds = Math.max(0, Math.floor(value));
	const minutes = Math.floor(totalSeconds / 60);
	const seconds = totalSeconds % 60;
	return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

function formatTimestampRange(start: number, end?: number) {
	const startLabel = formatTimestamp(start);
	if (!end || end <= start) {
		return startLabel;
	}
	return `${startLabel}-${formatTimestamp(end)}`;
}

function citationHref(citation: RAGCitation) {
	if (!citation.source_url) {
		return `/app/library/${citation.content_id}`;
	}
	if (citation.type !== "video" || !citation.timestamp) {
		return citation.source_url;
	}

	try {
		const url = new URL(citation.source_url);
		url.searchParams.set("t", `${Math.max(0, Math.floor(citation.timestamp.start))}s`);
		return url.toString();
	} catch {
		return citation.source_url;
	}
}
*/

