# Directional Stimulus Prompting (DSP)

**Source:** https://www.promptingguide.ai/techniques/dsp

## What it is
Guide a frozen black-box LLM toward a desired output by injecting a small
**stimulus / hint** into the prompt. A **small tunable policy model** is trained
(often via RL) to generate these hints; the large LLM stays frozen.

## When to use
- Steering outputs (e.g. summaries) toward specific content, keywords, or aspects
  without fine-tuning the big model.
- You can train/optimize a lightweight helper that emits hints.

## Concept
```
[policy LM] --> hint/keywords  ┐
                               ├──> [frozen LLM] --> desired output
user input --------------------┘
```
The hint nudges the LLM to include desired elements (e.g. must-mention keywords
for a summary).

## Practical (no-training) approximation
Manually provide the "stimulus": key points/keywords the answer must hit.
```
Summarize the article below. Make sure to include: <keyword1>, <keyword2>, <angle>.

<article>
```

## Pitfalls
- The full method needs training a policy model + RL; the manual hint version is
  a lightweight stand-in.

## Related
RAG, Generated-Knowledge, APE.
