# promptsmith

A CLI prompt optimizer. Hand it a rough prompt or task description and it returns
a **polished, technique-backed** version — powered by an LLM and grounded in the
17 proven prompt-engineering techniques from
[promptingguide.ai](https://www.promptingguide.ai/techniques) (zero-shot,
few-shot, chain-of-thought, self-consistency, ReAct, meta-prompting, ToT, RAG,
PAL, Reflexion, and more).

Works with any OpenAI-compatible endpoint. Defaults to a local
[OmniLLM](https://github.com/OmniLLM/omnillm) proxy at `http://localhost:5000/v1`.

## Install

```bash
git clone https://github.com/OmniLLM/promptsmith.git
cd promptsmith/promptsmith
./install.sh          # builds the Go binary into ~/.local/bin + scaffolds config
```

Or build manually:

```bash
cd promptsmith/promptsmith
go build -o promptsmith .
```

Written in Go — no runtime dependencies, single static binary. Requires Go >= 1.21 to build.

## Usage

```bash
promptsmith "write a tweet about cats"
echo "summarize this article" | promptsmith
promptsmith -m claude-opus-4.8 "classify sentiment of reviews"
promptsmith --raw "just give me the rewritten prompt"   # polished prompt only, no explanation
promptsmith --list-models                                # list available models
```

Default output is structured:

```
## Diagnosis          what's weak in your prompt
## Technique(s) applied  which techniques and why
## Polished prompt    the rewritten prompt, ready to paste
## Knobs to tune      shots / temperature / tools / retrieval source
```

Use `--raw` to get only the rewritten prompt (great for piping/scripting).

## Configuration

Precedence: **CLI flag > env var > config file > default**.

| Setting | Flag | Env var | Config key | Default |
|---|---|---|---|---|
| Base URL | `-u` | `PROMPTSMITH_BASE_URL` | `base_url` | `http://localhost:5000/v1` |
| Model | `-m` | `PROMPTSMITH_MODEL` | `model` | `gpt-5.5` |
| API key | `-k` | `PROMPTSMITH_API_KEY` | `api_key` | reads `~/.config/omnillm/api-key` |
| Temperature | `-t` | — | — | `0.3` |

Config file: `~/.config/promptsmith/config.json`

```json
{
  "base_url": "http://localhost:5000/v1",
  "model": "gpt-5.5"
}
```

Point it at OpenAI, Anthropic-via-proxy, a local vLLM, etc. by changing `base_url`
and `api_key`.

## Hermes skill

This repo also ships the original **`prompt-engineering`** Hermes Agent skill
(the methodology the CLI is built on), including a full reference guide per
technique under `prompt-engineering/references/`. Install it with:

```bash
ln -s "$PWD/prompt-engineering" ~/.hermes/skills/prompt-engineering
```

## Layout

```
promptsmith/
  promptsmith/            the CLI (Go)
    main.go               single-file Go CLI (stdlib only)
    go.mod
    install.sh            installer (go build + config scaffold)
  prompt-engineering/     the Hermes skill (methodology + 17 reference guides)
```
