# Reflexion

**Source:** https://www.promptingguide.ai/techniques/reflexion

## What it is
A framework for **verbal reinforcement**: an agent converts feedback from the
environment into **self-reflection** (linguistic feedback) stored in memory, and
uses it as context on the next attempt — learning from mistakes **without
fine-tuning** the model.

## Three components
- **Actor** — generates text/actions (built on CoT or ReAct) + a memory.
- **Evaluator** — scores the produced trajectory (reward).
- **Self-Reflection** — LLM turns the reward + trajectory into concrete verbal
  feedback, saved to long-term memory for the next trial.

Loop: define task → generate trajectory → evaluate → reflect → retry.

## When to use
- Agents that **learn by trial and error**: decision-making, reasoning,
  programming (retry failing code with reflections on why it failed).
- When RL fine-tuning is impractical and you want **nuanced, interpretable**
  feedback and explicit episodic memory.

## Pattern
```
Attempt 1: <agent trajectory> → FAIL (evaluator: <why>)
Reflection: <what went wrong and what to do differently, in words>
Attempt 2: <agent uses the reflection as context> → ...
```

## Polishing tips
- Keep reflections **specific and actionable** ("the test failed because I didn't
  handle empty input; add a guard").
- Store reflections in memory across trials (sliding window, or vector/SQL for
  long tasks).

## Pitfalls
- Relies on the model's ability to self-evaluate accurately.
- Long-term memory constraints; test-driven code has non-determinism limits.

## Related
ReAct + CoT (Actor base), Self-Consistency.
