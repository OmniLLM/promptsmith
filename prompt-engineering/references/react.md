# ReAct Prompting

**Source:** https://www.promptingguide.ai/techniques/react

## What it is
Interleave **Reasoning** and **Acting**: the model emits `Thought → Action →
Observation` steps in a loop. Thoughts plan/track/update; Actions call external
tools (search, calculator, APIs); Observations feed results back. Grounds
reasoning in the real world, reducing hallucination and enabling dynamic plans.

## When to use
- Tasks needing **external information or tools**: multi-hop QA, fact
  verification, calculations, browsing, agentic workflows.
- Anytime pure CoT would hallucinate facts it can't know.

## Pattern (few-shot exemplars, then the task)
```
Question: <question>
Thought 1: <what I need to figure out / next step>
Action 1: Search[<query>]
Observation 1: <tool result>
Thought 2: <reasoning over the observation>
Action 2: Lookup[<term>]
Observation 2: <result>
...
Thought N: <conclusion>
Action N: Finish[<final answer>]
```

Define your action space, e.g. `Search[]`, `Lookup[]`, `Calculator[]`, `Finish[]`.

## Polishing tips
- Provide 1–3 ReAct-format exemplars showing good Thought/Action/Observation.
- For reasoning-heavy tasks use many thought-action steps; for action-heavy
  decision tasks, thoughts can be sparse.
- Best results: **ReAct + CoT + Self-Consistency**, switching between internal
  reasoning and external retrieval.

## Pitfalls
- Depends heavily on tool/search quality; non-informative results derail it.
- Structural constraint can reduce reasoning flexibility.

## Related
CoT, Self-Consistency, ART, Reflexion (adds learning from failed trajectories).
