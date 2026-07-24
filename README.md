# omni-skills

A collection of commonly-used [Hermes Agent](https://hermes-agent.nousresearch.com) skills.

## Skills

| Skill | What it does |
|---|---|
| [`prompt-engineering`](./prompt-engineering) | Polish and strengthen prompts using 17 proven prompt-engineering techniques (zero-shot, few-shot, CoT, self-consistency, ReAct, meta-prompting, ToT, RAG, PAL, Reflexion, and more). Distilled from [promptingguide.ai](https://www.promptingguide.ai/techniques). Trigger it with "polish my prompt", "improve this prompt", or "which prompting technique should I use". |

## Layout

Each skill is a directory containing a `SKILL.md` (the entry point Hermes loads)
plus optional `references/`, `templates/`, `scripts/`, and `assets/`.

## Install

Clone into your Hermes skills directory, or symlink individual skills:

```bash
git clone https://github.com/OmniLLM/omni-skills.git
ln -s "$PWD/omni-skills/prompt-engineering" ~/.hermes/skills/prompt-engineering
```
