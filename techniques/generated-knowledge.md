# Generated Knowledge Prompting

**Source:** https://www.promptingguide.ai/techniques/knowledge

## What it is
Ask the model to **generate relevant facts/knowledge first**, then feed that
knowledge back in to answer the actual question. Helps commonsense and
knowledge-intensive tasks where the model answers wrong from lack of grounding.

## When to use
- The model gives a confidently wrong answer due to missing world knowledge.
- No external retrieval available, but the model *does* know the facts if prompted.

## Two-step pattern
**Step 1 — generate knowledge** (few-shot the format):
```
Input: <statement>
Knowledge: <a relevant factual paragraph>
...
Input: <your question's subject>
Knowledge:
```

**Step 2 — integrate & answer:**
```
Question: <the question>? Yes or No?
Knowledge: <generated knowledge from step 1>
Explain and Answer:
```

## Example
Q "Part of golf is trying to get a higher point total than others?" → naive Yes
(wrong). Generate knowledge about golf scoring → re-ask → correct "No".

## Polishing tips
- Generate **multiple** knowledge snippets and pick/ensemble; confidence varies.
- Reformat the final question into QA format to guide the answer shape.

## Pitfalls
- Generated "knowledge" can itself be wrong → for facts that must be correct,
  use **RAG** with a real source instead.

## Related
RAG (external retrieval), CoT.
