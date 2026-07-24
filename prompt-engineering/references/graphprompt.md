# GraphPrompt

**Source:** https://www.promptingguide.ai/techniques/graph

## What it is
A prompting framework for **graph-structured data** (Liu et al. 2023), designed
to improve performance on downstream graph tasks by unifying pre-training and
prompting over graphs. (The guide notes this section is brief — "more coming
soon".)

## When to use
- Tasks over graphs/networks: node classification, link prediction, graph
  classification, knowledge-graph reasoning.

## Practical guidance for everyday prompting
When you must reason over graph/relational data with a general LLM:
- **Serialize the graph clearly**: list nodes, then edges as `A --rel--> B`, or
  provide an adjacency list / triples `(subject, relation, object)`.
- Use consistent delimiters and IDs so the model can traverse relationships.
- Combine with **CoT** to walk paths step by step, and **RAG** to pull in
  relevant subgraphs.

## Pattern
```
Graph:
Nodes: A, B, C, D
Edges: A->B (friend), B->C (works_with), C->D (reports_to)

Question: <a path/relationship question>
Let's trace the relevant edges step by step.
```

## Pitfalls
- Large graphs exceed context; retrieve/prune to the relevant subgraph first.

## Related
CoT, RAG.
