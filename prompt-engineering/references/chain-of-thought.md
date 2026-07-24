# Chain-of-Thought (CoT) Prompting

**Source:** https://www.promptingguide.ai/techniques/cot

## What it is
Elicit **intermediate reasoning steps** before the final answer. Dramatically
improves arithmetic, commonsense, and symbolic reasoning. An emergent ability of
sufficiently large models.

## Variants
1. **Few-shot CoT** — demonstrations that *show the reasoning*, then the answer.
2. **Zero-shot CoT** (Kojima 2022) — just append **"Let's think step by step"**.
   No examples needed; great when you have none.
3. **Auto-CoT** (Zhang 2022) — auto-generate diverse reasoning chains: cluster
   questions, sample one per cluster, generate its chain with zero-shot CoT.

## Pattern — few-shot CoT
```
Q: <problem>
A: <step-by-step reasoning> ... The answer is <X>.

Q: <problem>
A: <step-by-step reasoning> ... The answer is <Y>.

Q: <your problem>
A:
```

## Pattern — zero-shot CoT
```
<your problem>

Let's think step by step.
```

## When to use
- Any task needing multiple reasoning steps: math word problems, logic,
  multi-hop questions, planning.

## Polishing tips
- Put the reasoning **before** the final answer (answering first invites
  rationalizing a wrong answer).
- Even **one** good worked example often suffices.
- End with a clear "The answer is ___" anchor to make parsing easy.

## Pitfalls
- Reasoning can still be confidently wrong → verify or add **Self-Consistency**
  (sample many chains, majority vote).
- On tiny models CoT may not help (emergent ability).

## Stacks well with
Few-shot, Self-Consistency, ReAct.
