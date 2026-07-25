# Automatic Reasoning and Tool-use (ART)

**Source:** https://www.promptingguide.ai/techniques/art

## What it is
Uses a frozen LLM to **automatically** generate intermediate reasoning steps as
a program, selecting multi-step reasoning + tool-use demonstrations from a
**task library**. At runtime it **pauses generation when a tool is called**,
runs the tool, integrates the output, and resumes — no hand-scripted interleaving.

## When to use
- You want ReAct-style reasoning+tools but without hand-crafting task-specific
  demonstrations for every new task.
- Unseen tasks that resemble ones in your library (generalizes zero-shot).

## How it works
1. New task → retrieve related demonstrations of multi-step reasoning + tool use
   from a **task library**.
2. Generate a reasoning program; **stop at tool calls**, run them, splice results.
3. Extensible: humans can fix reasoning steps or add tools by updating the task
   and tool libraries.

## Polishing tips
- Maintain a good library of decomposed exemplars and tools.
- Incorporate human feedback to correct reasoning — exceeds hand-crafted CoT
  when feedback is added.

## Pitfalls
- Needs infrastructure (library + tool runtime); heavier than plain ReAct.

## Related
ReAct, PAL, Automatic Prompt Engineer (APE).
