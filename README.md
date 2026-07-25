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

## Configuring the LLM provider

promptsmith speaks the **OpenAI Chat Completions API** (`POST /chat/completions`
and `GET /models`). Any provider or proxy that exposes that shape works — you
only need to set three things: **base URL**, **model**, and **API key**.

### Settings & precedence

Every setting can come from a CLI flag, an environment variable, the config
file, or a built-in default. Higher in this list wins:

**CLI flag > env var > config file > default**

| Setting | Flag | Env var | Config key | Default |
|---|---|---|---|---|
| Base URL | `-u`, `--base-url` | `PROMPTSMITH_BASE_URL` | `base_url` | `http://localhost:5000/v1` |
| Model | `-m`, `--model` | `PROMPTSMITH_MODEL` | `model` | `gpt-5.5` |
| API key | `-k`, `--api-key` | `PROMPTSMITH_API_KEY` | `api_key` | falls back to `~/.config/omnillm/api-key` |
| Temperature | `-t`, `--temperature` | — | — | `0.3` |

> The base URL must include the API version path (e.g. `.../v1`). promptsmith
> appends `/chat/completions` and `/models` to whatever you give it.

### Config file

`~/.config/promptsmith/config.json` (created by `install.sh`):

```json
{
  "base_url": "http://localhost:5000/v1",
  "model": "gpt-5.5",
  "api_key": ""
}
```

Leave `api_key` empty to use an env var or the OmniLLM key file instead. Keeping
the key out of the config file and in `PROMPTSMITH_API_KEY` is recommended.

### Provider recipes

**OmniLLM (default, local proxy)** — nothing to configure. promptsmith points at
`http://localhost:5000/v1` and reads the key from `~/.config/omnillm/api-key`
automatically. Just run `promptsmith "..."`.

**OpenAI**

```json
{
  "base_url": "https://api.openai.com/v1",
  "model": "gpt-4o"
}
```
```bash
export PROMPTSMITH_API_KEY="sk-..."
promptsmith "polish this prompt"
```

**Anthropic (via an OpenAI-compatible gateway/proxy)** — Anthropic's native API
is not OpenAI-shaped, so route it through a proxy such as OmniLLM, LiteLLM, or
one-api:

```json
{ "base_url": "http://localhost:5000/v1", "model": "claude-opus-4.8" }
```

**Groq**

```json
{ "base_url": "https://api.groq.com/openai/v1", "model": "llama-3.3-70b-versatile" }
```
```bash
export PROMPTSMITH_API_KEY="gsk_..."
```

**OpenRouter**

```json
{ "base_url": "https://openrouter.ai/api/v1", "model": "anthropic/claude-3.5-sonnet" }
```
```bash
export PROMPTSMITH_API_KEY="sk-or-..."
```

**Local vLLM / llama.cpp / Ollama** — anything exposing an OpenAI-compatible
server:

```bash
# vLLM
promptsmith -u http://localhost:8000/v1 -m meta-llama/Llama-3.1-8B-Instruct -k EMPTY "..."
# Ollama
promptsmith -u http://localhost:11434/v1 -m llama3.1 -k ollama "..."
```

Local servers usually ignore the key, but the OpenAI client still requires a
non-empty value — pass any placeholder (`EMPTY`, `ollama`, etc.).

### One-off overrides

Flags override everything for a single run without touching your config:

```bash
promptsmith -u https://api.openai.com/v1 -m gpt-4o -k "$OPENAI_API_KEY" "improve this"
```

### Verify your setup

```bash
promptsmith --list-models     # confirms base URL + key reach the provider
```

If this prints models, you're wired up correctly. A connection or `HTTP 401`
error here points to a wrong `base_url` or `api_key`.

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
