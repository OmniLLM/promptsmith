# Multimodal CoT Prompting

**Source:** https://www.promptingguide.ai/techniques/multimodalcot

## What it is
Extends chain-of-thought to **text + vision**. A two-stage framework:
1. **Rationale generation** — produce reasoning grounded in *both* the image and
   text.
2. **Answer inference** — use the generated rationale to produce the final answer.

Traditional CoT is language-only; Multimodal CoT incorporates the image so the
reasoning can reference visual content. A 1B multimodal-CoT model outperformed
GPT-3.5 on ScienceQA.

## When to use
- Questions that require reasoning over an **image + text** (diagrams, charts,
  science questions with figures, visual QA needing steps).

## Pattern
```
[Image] + Question
Stage 1 → Rationale: <reasoning that cites what's in the image and the text>
Stage 2 → Answer: <final answer using the rationale>
```

## Polishing tips
- Ask the model to **describe the relevant visual evidence first**, then reason,
  then answer.
- Keep the rationale explicitly tied to image features.

## Pitfalls
- Requires a vision-capable model.
- Rationale quality gates the answer; a wrong visual read propagates.

## Related
CoT, Generated-Knowledge.
