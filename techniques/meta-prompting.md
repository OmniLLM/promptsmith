# Meta Prompting

**Source:** https://www.promptingguide.ai/techniques/meta-prompting

## What it is
An advanced technique focusing on the **structure and syntax** of a problem and
its solution rather than specific content. You give the model an **abstract
template/skeleton** of how to solve a class of problems, not concrete examples.

## Key characteristics (Zhang et al. 2024)
1. **Structure-oriented** — format/pattern over content.
2. **Syntax-focused** — syntax as a template for the response.
3. **Abstract examples** — frameworks illustrating structure without details.
4. **Versatile** — one structure covers many instances of a problem type.
5. **Categorical** — logical arrangement of prompt components (type theory).

## Advantages over few-shot
- **Token-efficient** — no long concrete examples.
- **Fairer comparison** — minimizes bias from specific examples.
- **Zero-shot-like** — reduces influence of particular exemplars.

## When to use
- Complex reasoning, math, coding challenges, theoretical queries where a
  **reusable solution skeleton** helps and you want to save tokens.
- The model already has innate knowledge of the task type.

## Pattern
```
To solve problems of type <T>, follow this structure:
1. <abstract step: identify given/unknown>
2. <abstract step: choose method>
3. <abstract step: derive>
4. <abstract step: state result as ...>

Now solve: <concrete problem>
```

## Pitfalls
- Assumes innate task knowledge — degrades on novel/unique tasks (like zero-shot).

## Related
Few-shot (contrast), CoT, Zero-shot.
