# Self-Consistency

**Source:** https://www.promptingguide.ai/techniques/consistency

## What it is
Replace greedy decoding in CoT with **sampling multiple diverse reasoning paths**
(higher temperature), then **take the majority-vote answer**. Boosts arithmetic
and commonsense reasoning by averaging out any single bad chain.

## When to use
- CoT gives inconsistent or occasionally wrong answers.
- Accuracy matters more than token cost.
- The final answer is a discrete value you can vote on (number, label, choice).

## How to apply
1. Build a few-shot **CoT** prompt.
2. Sample the model **N times** (e.g. 5–40) with temperature > 0.
3. Extract the final answer from each sample.
4. **Return the most frequent answer.**

## Pattern
Same CoT prompt, run repeatedly:
```
Q: When I was 6 my sister was half my age. Now I'm 70. How old is my sister?
A: <reasoning> ... The answer is 67.
```
Sample 10× → majority = 67 (vs a single greedy run that said 35).

## Polishing tips
- Temperature ~0.5–0.7 for diverse-but-sane chains.
- Odd N avoids ties; break ties by highest-confidence chain.
- Only worth it when a single CoT is unreliable.

## Pitfalls
- N× the cost/latency. Don't use for simple deterministic tasks.
- Needs an extractable, comparable final answer; free-form text is hard to vote on.

## Stacks well with
CoT (required base), ReAct (switch between ReAct and CoT+Self-Consistency).
