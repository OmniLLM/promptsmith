# Retrieval Augmented Generation (RAG)

**Source:** https://www.promptingguide.ai/techniques/rag

## What it is
Combine an **information-retrieval component** with the generator: retrieve
relevant documents from an external source (vector index, DB, web), put them in
the prompt as context, then generate the answer. Improves factual accuracy and
**reduces hallucination**; knowledge can be updated without retraining.

## When to use
- Knowledge-intensive tasks needing **current, private, or domain** facts.
- You must ground answers in citable sources.
- Facts change over time (docs, policies, prices, news).

## Pattern
```
Use ONLY the context below to answer. If the answer isn't in the context, say
you don't know.

### Context
<retrieved chunk 1>
<retrieved chunk 2>

### Question
<user question>
```

## Pipeline
1. Chunk + embed your knowledge base into a vector index.
2. Embed the query; retrieve top-k relevant chunks.
3. Insert chunks as context; instruct the model to answer *from context only*.
4. (Optional) ask for citations to the chunks used.

## Polishing tips
- Tell the model to **refuse when context is insufficient** (kills hallucination).
- Keep chunks focused; rerank for relevance; include source IDs for citation.
- Balance k vs. context window and noise.

## Pitfalls
- Garbage retrieval → garbage answer. Retrieval quality is the bottleneck.
- Conflicting chunks confuse the model; dedupe/rerank.

## Related
Generated-Knowledge (self-sourced facts), ReAct (retrieve as an action).
