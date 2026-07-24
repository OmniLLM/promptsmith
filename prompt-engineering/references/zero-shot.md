# Zero-Shot Prompting

**Source:** https://www.promptingguide.ai/techniques/zeroshot

## What it is
Instruct the model to do a task with **no examples**. Modern instruction-tuned
LLMs (via instruction tuning + RLHF) perform many tasks directly from a clear
instruction. Always the first thing to try — cheapest, simplest.

## When to use
- Task is common and the model likely already "knows" it (sentiment, translation,
  simple extraction, generic Q&A).
- You want a baseline before adding complexity.

## Pattern
```
<clear instruction describing the task and desired output>

<input>
<format anchor>:
```

## Example
```
Classify the text into neutral, negative or positive.

Text: I think the vacation is okay.
Sentiment:
```
→ `Neutral`

## Polishing tips
- State the task, constraints, and **output format** explicitly.
- Add a format anchor (`Sentiment:`, `JSON:`) — steers the shape of the answer.
- Assign a role/persona if it sharpens tone ("You are a senior tax advisor").
- For reasoning tasks, combine with **zero-shot CoT** ("Let's think step by
  step") — see `chain-of-thought.md`.

## Pitfalls
- Fails on complex multi-step reasoning → escalate to few-shot or CoT.
- Ambiguous instructions produce inconsistent output; nail down format + scope.

## Escalate when it fails
Few-shot → CoT → Self-Consistency.
