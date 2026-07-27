# Few-Shot Prompting

**Source:** https://www.promptingguide.ai/techniques/fewshot

## What it is
Provide a handful of **demonstrations** (input→output pairs) in the prompt so the
model learns the task in-context. Enables tasks zero-shot can't reliably do.
1-shot, 3-shot, 5-shot... increase until it works.

## When to use
- You need a specific output format, label space, or style.
- Zero-shot is flaky or inconsistent.
- The task has a recognizable pattern examples convey faster than rules.

## Pattern
```
<example 1 input> // <label 1>
<example 2 input> // <label 2>
<example 3 input> // <label 3>
<your input> //
```

## Research-backed tips (Min et al. 2022)
- **Label space and input distribution matter** — even *random* labels beat no
  labels, because the model learns the shape of the task.
- **Format matters** — keep a consistent, clear format across all examples.
- Draw example labels from the **true distribution**, not uniform.
- Pick **diverse, representative** examples; avoid leaking the test answer.

## Polishing tips
- 2–8 examples is the usual sweet spot; more for harder tasks.
- Keep example formatting identical to what you want back.
- Order can matter — put a clean, unambiguous example last.

## Pitfalls
- Standard few-shot still fails on **multi-step reasoning** (e.g. "do the odd
  numbers add to an even number?"). → add reasoning: **Chain-of-Thought**.
- Token-heavy; if cost matters, consider **meta-prompting** (structure only).

## Escalate when it fails
Few-shot + CoT → Self-Consistency → fine-tuning.
