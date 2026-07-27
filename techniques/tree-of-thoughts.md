# Tree of Thoughts (ToT)

**Source:** https://www.promptingguide.ai/techniques/tot

## What it is
Generalizes CoT into a **search over a tree of intermediate "thoughts"**. The
model generates candidate thoughts, **self-evaluates** them, and uses search
(BFS/DFS/beam) with **lookahead and backtracking** to solve problems needing
exploration or strategic planning.

## When to use
- Complex problems where a single linear chain gets stuck: puzzles, Game of 24,
  planning, constraint search, creative multi-step design.

## Framework knobs
- **#candidates** per step, **#steps** to decompose into.
- Evaluation prompt: rate each partial thought "sure / maybe / impossible".
- Search strategy: BFS keeps best `b` candidates per level; DFS backtracks.

## Simple single-prompt variant (Hulbert 2023)
Great for everyday use without building a search loop:
```
Imagine three different experts are answering this question.
All experts will write down 1 step of their thinking, then share it with the group.
Then all experts go on to the next step, etc.
If any expert realises they're wrong at any point then they leave.
The question is: <your question>
```
(PanelGPT / "panel discussion" is a related variant.)

## Polishing tips
- Define steps explicitly and ask the model to **evaluate** each partial solution
  before continuing.
- Prune "impossible" branches early; keep "maybe" alive.

## Pitfalls
- Full ToT needs orchestration code (generate→evaluate→search); the 3-experts
  prompt approximates it in one shot but is weaker.
- Token-heavy.

## Related
CoT (base), Self-Consistency, ReAct.
