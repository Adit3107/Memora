/*
=============================================================================
PAUSED FEATURE: AI Playground (Put on hold for upcoming release)
Full interactive conversational RAG playground with scope selection & citations.
Separated and commented out so the developer can focus on working code.
To resume: uncomment this file and re-mount in src/app/app/ai/page.tsx.
=============================================================================

"use client";

import {
	Bot,
	Check,
	FileText,
	Library,
	MessagesSquare,
	Send,
	Sparkles,
	Video,
} from "lucide-react";
import { FormEvent, useEffect, useMemo, useState } from "react";

import {
	askMemora,
	displayContentName,
	listContent,
	MEMORA_DEMO_USER_ID,
	type BackendContent,
	type RAGCitation,
	type RAGMessage,
	type RAGScope,
} from "@/lib/api";

type PlaygroundMode = "library" | "content" | "selected_sources";

type ThreadMessage = RAGMessage & {
	citations?: RAGCitation[];
};

const modes: {
	value: PlaygroundMode;
	label: string;
	icon: typeof Library;
}[] = [
	{ value: "library", label: "Library", icon: Library },
	{ value: "content", label: "This Content", icon: FileText },
	{ value: "selected_sources", label: "Selected", icon: Check },
];

const modeSuggestions: Record<PlaygroundMode, string[]> = {
	library: [
		"What did I save about Kafka?",
		"What topics appear in my library?",
		"Summarize my saved knowledge.",
	],
	content: [
		"What is this content about?",
		"Explain the important concepts.",
		"What are the main takeaways?",
	],
	selected_sources: [
		"Compare these sources.",
		"What do these sources agree on?",
		"Summarize the selected items.",
	],
};

export function AIPlayground() {
	const [mode, setMode] = useState<PlaygroundMode>("library");
	const [contents, setContents] = useState<BackendContent[]>([]);
	const [selectedContentID, setSelectedContentID] = useState("");
	const [selectedContentIDs, setSelectedContentIDs] = useState<string[]>([]);
	const [question, setQuestion] = useState("");
	const [messages, setMessages] = useState<ThreadMessage[]>([]);
	const [isAsking, setIsAsking] = useState(false);
	const [isLoadingContent, setIsLoadingContent] = useState(true);
	const [error, setError] = useState("");

	async function loadContent() {
		setIsLoadingContent(true);
		try {
			const rows = await listContent();
			setContents(rows);
			setSelectedContentID((current) => current || rows[0]?.id || "");
		} catch {
			setError("Content could not be loaded. Library mode is still available.");
		} finally {
			setIsLoadingContent(false);
		}
	}

	useEffect(() => {
		// eslint-disable-next-line react-hooks/set-state-in-effect
		void loadContent();
	}, []);

	const scope = useMemo((): RAGScope => {
		if (mode === "content") {
			return { type: "content", content_id: selectedContentID };
		}
		if (mode === "selected_sources") {
			return { type: "selected_sources", content_ids: selectedContentIDs };
		}
		return { type: "library" };
	}, [mode, selectedContentID, selectedContentIDs]);

	const canAsk =
		!isAsking &&
		question.trim().length > 0 &&
		(mode === "library" ||
			(mode === "content" && selectedContentID) ||
			(mode === "selected_sources" && selectedContentIDs.length > 0));

	function setModeAndReset(nextMode: PlaygroundMode) {
		setMode(nextMode);
		setError("");
	}

	function toggleSelectedSource(contentID: string) {
		setSelectedContentIDs((current) =>
			current.includes(contentID)
				? current.filter((candidate) => candidate !== contentID)
				: [...current, contentID]
		);
	}

	async function submitQuestion(event?: FormEvent<HTMLFormElement>, value = question) {
		event?.preventDefault();
		const trimmedQuestion = value.trim();
		if (
			!trimmedQuestion ||
			isAsking ||
			(mode === "content" && !selectedContentID) ||
			(mode === "selected_sources" && selectedContentIDs.length === 0)
		) {
			return;
		}

		const history = messages
			.slice(-6)
			.map((message) => ({ role: message.role, content: message.content }));

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
			setError("MEMORA could not answer that question. Check the backend and AI services.");
		} finally {
			setIsAsking(false);
		}
	}

	return (
		<div className="grid min-h-[calc(100vh-9rem)] gap-4 xl:grid-cols-[320px_1fr]">
			<aside className="rounded-md border bg-card p-4 shadow-sm">
				<div className="flex items-center gap-3">
					<span className="flex size-10 items-center justify-center rounded-md border bg-background">
						<Bot className="size-5 text-primary" aria-hidden="true" />
					</span>
					<div>
						<p className="text-sm font-medium uppercase text-muted-foreground">
							AI Playground
						</p>
						<h1 className="text-xl font-semibold">Ask MEMORA</h1>
					</div>
				</div>

				<div className="mt-5 grid gap-2">
					{modes.map((item) => {
						const Icon = item.icon;
						const isActive = mode === item.value;
						return (
							<button
								className={[
									"flex h-11 items-center gap-3 rounded-md border px-3 text-sm font-medium transition-colors",
									isActive
										? "bg-primary text-primary-foreground"
										: "bg-background text-muted-foreground hover:bg-accent hover:text-accent-foreground",
								].join(" ")}
								key={item.value}
								onClick={() => setModeAndReset(item.value)}
								type="button"
							>
								<Icon className="size-4" aria-hidden="true" />
								{item.label}
							</button>
						);
					})}
				</div>

				<div className="mt-5">
					{mode === "content" ? (
						<label className="space-y-2 text-sm font-medium">
							Source
							<select
								className="h-11 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
								disabled={isLoadingContent || contents.length === 0}
								onChange={(event) => setSelectedContentID(event.target.value)}
								value={selectedContentID}
							>
								{contents.map((item) => (
									<option key={item.id} value={item.id}>
										{displayContentName(item)}
									</option>
								))}
							</select>
						</label>
					) : null}

					{mode === "selected_sources" ? (
						<div className="space-y-3">
							<div className="flex items-center justify-between gap-3">
								<p className="text-sm font-medium">Sources</p>
								<button
									className="text-xs font-medium text-muted-foreground hover:text-foreground disabled:opacity-50"
									disabled={selectedContentIDs.length === 0}
									onClick={() => setSelectedContentIDs([])}
									type="button"
								>
									Clear
								</button>
							</div>
							<div className="max-h-80 space-y-2 overflow-y-auto pr-1">
								{contents.map((item) => (
									<label
										className="flex cursor-pointer items-start gap-3 rounded-md border bg-background p-3 text-sm transition-colors hover:bg-accent"
										key={item.id}
									>
										<input
											checked={selectedContentIDs.includes(item.id)}
											className="mt-1 size-4"
											onChange={() => toggleSelectedSource(item.id)}
											type="checkbox"
										/>
										<span className="min-w-0">
											<span className="block truncate font-medium">{displayContentName(item)}</span>
											<span className="block text-xs text-muted-foreground">
												{item.type}
											</span>
										</span>
									</label>
								))}
							</div>
						</div>
					) : null}
				</div>
			</aside>

			<section className="flex min-h-0 flex-col rounded-md border bg-card shadow-sm">
				<div className="flex items-center justify-between gap-3 border-b px-4 py-3">
					<div className="flex items-center gap-3">
						<span className="flex size-9 items-center justify-center rounded-md border bg-background">
							<MessagesSquare className="size-4 text-primary" aria-hidden="true" />
						</span>
						<div>
							<h2 className="text-base font-semibold">{modeTitle(mode)}</h2>
							<p className="text-xs text-muted-foreground">{modeStatus(mode, contents, selectedContentIDs)}</p>
						</div>
					</div>
					<button
						className="rounded-md border bg-background px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground disabled:opacity-50"
						disabled={messages.length === 0 || isAsking}
						onClick={() => {
							setMessages([]);
							setError("");
						}}
						type="button"
					>
						New chat
					</button>
				</div>

				<div className="flex-1 overflow-y-auto p-4">
					{messages.length === 0 ? (
						<div className="flex min-h-[420px] flex-col items-center justify-center text-center">
							<div className="flex size-14 items-center justify-center rounded-md border bg-background">
								<Sparkles className="size-6 text-primary" aria-hidden="true" />
							</div>
							<h2 className="mt-5 text-2xl font-semibold">Start a memory chat</h2>
							<div className="mt-5 flex max-w-2xl flex-wrap justify-center gap-2">
								{modeSuggestions[mode].map((suggestion) => (
									<button
										className="rounded-md border bg-background px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground disabled:opacity-50"
										disabled={!canAskForSuggestion(mode, selectedContentID, selectedContentIDs)}
										key={suggestion}
										onClick={() => void submitQuestion(undefined, suggestion)}
										type="button"
									>
										{suggestion}
									</button>
								))}
							</div>
						</div>
					) : (
						<div className="mx-auto max-w-4xl space-y-4">
							{messages.map((message, index) => (
								<MessageBubble key={`${message.role}-${index}`} message={message} />
							))}
							{isAsking ? (
								<div className="rounded-md border bg-secondary/50 p-4 text-sm text-muted-foreground">
									Thinking...
								</div>
							) : null}
						</div>
					)}
				</div>

				{error ? <p className="border-t px-4 py-3 text-sm text-destructive">{error}</p> : null}

				<form className="border-t p-4" onSubmit={submitQuestion}>
					<div className="mx-auto flex max-w-4xl flex-col gap-3 sm:flex-row">
						<input
							className="h-12 flex-1 rounded-md border bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
							disabled={isAsking}
							onChange={(event) => setQuestion(event.target.value)}
							placeholder={inputPlaceholder(mode)}
							value={question}
						/>
						<button
							className="inline-flex h-12 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
							disabled={!canAsk}
							type="submit"
						>
							<Send className="size-4" aria-hidden="true" />
							Ask
						</button>
					</div>
				</form>
			</section>
		</div>
	);
}

function MessageBubble({ message }: { message: ThreadMessage }) {
	const isUser = message.role === "user";
	return (
		<div className={["flex", isUser ? "justify-end" : "justify-start"].join(" ")}>
			<div
				className={[
					"max-w-[88%] rounded-md border p-4",
					isUser ? "bg-primary text-primary-foreground" : "bg-secondary/50",
				].join(" ")}
			>
				<p className="whitespace-pre-wrap text-sm leading-6">{message.content}</p>
				{!isUser && message.citations && message.citations.length > 0 ? (
					<CitationList citations={message.citations} />
				) : null}
			</div>
		</div>
	);
}

function CitationList({ citations }: { citations: RAGCitation[] }) {
	return (
		<div className="mt-4 border-t pt-3">
			<p className="text-xs font-medium uppercase text-muted-foreground">Sources</p>
			<div className="mt-2 grid gap-2 sm:grid-cols-2">
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
							className="flex min-w-0 items-start gap-3 rounded-md border bg-background p-3 text-sm text-foreground transition-colors hover:bg-accent"
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

function modeTitle(mode: PlaygroundMode) {
	if (mode === "content") {
		return "Ask About This";
	}
	if (mode === "selected_sources") {
		return "Ask Selected Sources";
	}
	return "Ask My Library";
}

function modeStatus(
	mode: PlaygroundMode,
	contents: BackendContent[],
	selectedContentIDs: string[]
) {
	if (mode === "content") {
		return contents.length > 0 ? "One saved item" : "No saved items loaded";
	}
	if (mode === "selected_sources") {
		return `${selectedContentIDs.length} selected`;
	}
	return "Entire library";
}

function inputPlaceholder(mode: PlaygroundMode) {
	if (mode === "content") {
		return "Ask about the selected content...";
	}
	if (mode === "selected_sources") {
		return "Ask across the selected sources...";
	}
	return "Ask your library...";
}

function canAskForSuggestion(
	mode: PlaygroundMode,
	selectedContentID: string,
	selectedContentIDs: string[]
) {
	return (
		mode === "library" ||
		(mode === "content" && Boolean(selectedContentID)) ||
		(mode === "selected_sources" && selectedContentIDs.length > 0)
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

