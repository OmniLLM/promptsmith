# Active-Prompt

**Source:** https://www.promptingguide.ai/techniques/activeprompt

## What it is
CoT normally uses a **fixed set of human-annotated exemplars**, which may not be
optimal per task. Active-Prompt (Diao et al. 2023) selects **which questions are
most worth annotating** by measuring model **uncertainty**, then has humans
annotate those with CoT reasoning.

## How it works
1. Query the LLM (with or without a few CoT examples) **k times** per training
   question → k possible answers.
2. Compute an **uncertainty metric** (e.g. disagreement among the k answers).
3. Select the **most uncertain** questions for **human annotation** with CoT.
4. Use these new, high-value exemplars to prompt the remaining questions.

## When to use
- Building a strong few-shot CoT prompt for a dataset/task and you have limited
  human-annotation budget — spend it where the model is most confused.

## Polishing tips
- Use disagreement/entropy across k samples as the uncertainty score.
- Annotate a small, high-uncertainty set rather than random examples.

## Pitfalls
- Requires a pool of task questions and some human effort to annotate.

## Related
Few-shot, CoT, APE.
