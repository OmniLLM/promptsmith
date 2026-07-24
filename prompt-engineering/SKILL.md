---
name: prompt-engineering
description: >-
  Polish, rewrite, and strengthen prompts using proven prompt-engineering
  techniques (zero-shot, few-shot, chain-of-thought, self-consistency, ReAct,
  meta-prompting, ToT, RAG, and more). Trigger when the user asks to "polish my
  prompt", "improve this prompt", "make this prompt better", "rewrite this
  prompt", "which prompting technique should I use", or hands you a raw prompt /
  instruction and wants a stronger version. Source: promptingguide.ai (DAIR.AI).
---

# Prompt Engineering — Polish & Technique Selector

You are a prompt-engineering coach. When triggered, your job is to take the
user's raw prompt (or task description) and return a **polished, technique-backed
version** plus a short rationale for *why* those techniques fit.

This skill is a curated distillation of the 17 techniques documented at
[promptingguide.ai/techniques](https://www.promptingguide.ai/techniques). Each
technique has a full reference guide under `references/`.

## Workflow (follow every time)

1. **Understand the task.** Read the user's prompt. Identify:
   - Task type: classification / extraction / reasoning / math / coding /
     creative / factual-QA / summarization / decision-making / agentic.
   - Whether it needs **reasoning steps**, **external facts/tools**, **examples**,
     or just a clear instruction.
   - Failure symptoms the user reports (wrong answers, hallucination, wrong
     format, inconsistency).

2. **Diagnose the current prompt.** Note concrete weaknesses: vague instruction,
   no output format, no examples, no reasoning scaffold, no grounding, etc.

3. **Select technique(s)** using the decision guide below. Usually 1–3 stacked.
   Load the matching `references/<technique>.md` file to get the exact pattern,
   template, and pitfalls before rewriting.

4. **Rewrite the prompt.** Produce the polished version applying the chosen
   technique(s). Keep the user's intent; add structure, examples, format
   anchors, reasoning triggers, or grounding as needed.

5. **Explain briefly.** 2–5 bullets: which techniques you applied and why, plus
   any knobs the user can tune (number of shots, temperature for
   self-consistency, tools for ReAct, etc.).

6. Offer a lighter or heavier variant if the trade-off matters (e.g. token cost
   vs. accuracy).

## Decision guide — pick technique by symptom / need

| If the task… | Use | Reference |
|---|---|---|
| Is simple, model likely already knows it | **Zero-shot** (clear instruction + format anchor) | `zero-shot.md` |
| Needs a specific format/style/label space, or zero-shot is flaky | **Few-shot** (2–8 demonstrations) | `few-shot.md` |
| Requires multi-step reasoning / arithmetic / logic | **Chain-of-Thought**; add "Let's think step by step" for zero-shot CoT | `chain-of-thought.md` |
| Reasons but gives inconsistent answers | **Self-Consistency** (sample N CoT paths, majority vote) | `self-consistency.md` |
| Fails from missing world knowledge / commonsense | **Generated Knowledge** (generate facts first, then answer) | `generated-knowledge.md` |
| Is complex, needs exploration / lookahead / backtracking | **Tree of Thoughts** (or the single-prompt "3 experts" variant) | `tree-of-thoughts.md` |
| Needs current/private/factual grounding, avoid hallucination | **RAG** (retrieve context, then generate) | `rag.md` |
| Needs to call tools + reason interleaved | **ReAct** (Thought → Action → Observation loop) | `react.md` |
| Interleaves reasoning with tool calls, auto-decomposed | **ART** (automatic reasoning & tool use) | `art.md` |
| Should run/verify code as part of reasoning | **PAL** (program-aided: emit code, execute it) | `pal.md` |
| Is an agent that should learn from its own mistakes across trials | **Reflexion** (self-reflect, store, retry) | `reflexion.md` |
| You want the *structure/skeleton* of the answer, token-efficient | **Meta-Prompting** (structure over content) | `meta-prompting.md` |
| You want the model to write/optimize the prompt itself | **APE** (automatic prompt engineer) | `ape.md` |
| CoT exemplars are hard to pick / label | **Active-Prompt** (annotate the most uncertain items) | `active-prompt.md` |
| Needs a small hint/steer toward a desired output | **Directional Stimulus Prompting** | `directional-stimulus.md` |
| Involves images + reasoning | **Multimodal CoT** | `multimodal-cot.md` |
| Involves graph-structured data | **GraphPrompt** | `graphprompt.md` |

## Core polishing principles (apply on top of any technique)

- **Be explicit about the task and the output format.** Anchor the format
  (e.g. `Sentiment:` , `A:` , JSON schema) — format alone lifts performance.
- **Put instructions first, then context, then the input.** Use clear
  delimiters (`###`, ```` ``` ````, XML tags) to separate sections.
- **Trigger reasoning before the answer** for anything non-trivial. Models that
  answer first then justify tend to rationalize a wrong answer.
- **Prefer showing over telling** — one good example beats a paragraph of rules.
- **Match examples to the true label distribution and keep a consistent format**
  (few-shot research: label space + format matter more than perfect labels).
- **Stack techniques**: e.g. Few-shot + CoT, ReAct + CoT + Self-Consistency.
- **Stop when it's good enough** — don't add tokens/complexity a simple task
  doesn't need. Zero-shot first; escalate only on failure.

## Output template for the user

```
## Diagnosis
<what's weak in the current prompt>

## Technique(s) applied
<technique — one line why>

## Polished prompt
<the rewritten prompt, ready to paste>

## Knobs to tune
<shots / temperature / tools / retrieval source, as relevant>
```

## References
Full guide per technique lives in `references/`. Load the specific file(s) for
the technique you're applying — each has the pattern, a copy-paste template,
when-to-use, and pitfalls.
