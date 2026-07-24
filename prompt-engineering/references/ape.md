# Automatic Prompt Engineer (APE)

**Source:** https://www.promptingguide.ai/techniques/ape

## What it is
Frame instruction writing as **black-box optimization**: an LLM **generates
candidate instructions** for a task (from input/output demonstrations), a target
model executes each, and the best-scoring instruction is selected. The model
engineers the prompt for you.

## When to use
- You have example input→output pairs and want the **best instruction** without
  hand-tuning.
- Optimizing a prompt you'll reuse at scale.

## How it works
1. Give an inference LLM output demonstrations → it proposes instruction candidates.
2. Execute each candidate with a target model.
3. Score candidates on a held-out set; keep the best.

## Notable result
APE discovered a better zero-shot CoT trigger than "Let's think step by step":
> **"Let's work this out in a step by step way to be sure we have the right answer."**
which improved MultiArith and GSM8K.

## Related optimization methods
- **OPRO** — LLMs optimize prompts ("Take a deep breath" boosted math).
- **Prompt-OIRL** — offline inverse RL for query-dependent prompts.
- **AutoPrompt** — gradient-guided prompt search.
- **Prefix/Prompt Tuning** — learn soft (continuous) prompts via backprop.

## Practical DIY version
Ask the model: *"Here are 5 input/output examples. Propose 5 candidate
instructions that would produce these. Then critique and pick the best."*

## Pitfalls
- Needs an eval signal/dataset to score candidates.

## Related
Meta-Prompting, Active-Prompt, ART.
