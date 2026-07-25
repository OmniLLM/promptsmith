# PAL — Program-Aided Language Models

**Source:** https://www.promptingguide.ai/techniques/pal

## What it is
Instead of reasoning in free text, the model **emits code** (e.g. Python) as the
reasoning chain; an **interpreter executes** it to produce the answer. Offloads
exact computation/logic to a runtime, avoiding arithmetic mistakes.

## When to use
- Math, dates, unit conversion, data manipulation, anything where the model can
  *describe* the steps but fumbles the actual computation.
- Structured problems expressible as a short program.

## Pattern (few-shot with code-as-reasoning)
```
# Q: <problem>
# <comment reasoning>
<python statements>
answer = <expression>
```
Then run the emitted code and print `answer`.

## Example (date reasoning)
Prompt shows worked examples using `datetime`/`relativedelta`; model emits:
```
today = datetime(2023, 2, 27)
born = today - relativedelta(years=25)
born.strftime('%m/%d/%Y')   # -> 02/27/1998
```
Execute → correct date. The LLM plans; the interpreter computes.

## Polishing tips
- Give few-shot examples where **comments = reasoning** and code = steps.
- Constrain to a safe, sandboxed interpreter; `exec` untrusted output carefully.
- Ask for a single final variable to read back.

## Pitfalls
- Executing model-generated code is a security risk — sandbox it.
- Model may emit buggy code; validate/catch exceptions.

## Related
CoT (text reasoning), ReAct (code as an action/tool), ART.
