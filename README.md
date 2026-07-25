# promptsmith

A CLI prompt optimizer. Hand it a rough prompt or task description and it returns
a **polished, technique-backed** version — powered by an LLM and grounded in the
17 proven prompt-engineering techniques from
[promptingguide.ai](https://www.promptingguide.ai/techniques) (zero-shot,
few-shot, chain-of-thought, self-consistency, ReAct, meta-prompting, ToT, RAG,
PAL, Reflexion, and more).

Supports the same providers as
[omni-agent-desktop](https://github.com/OmniLLM/omni-agent-desktop) — a custom
provider in three API shapes (openai-compatible, anthropic-messages,
openai-responses) plus Azure AI Foundry. Defaults to a local
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

promptsmith supports the **same providers and wire shapes as
[omni-agent-desktop](https://github.com/OmniLLM/omni-agent-desktop)**, so a
config that works there works here. Two provider types, and the custom provider
speaks one of three API shapes:

| Provider (`-p`) | API shape (`-s`) | Request | Auth header |
|---|---|---|---|
| `custom` | `openai-compatible` (default) | `POST <base>/chat/completions` | `Authorization: Bearer` |
| `custom` | `anthropic-messages` | `POST <base>/messages` | `x-api-key` + `anthropic-version` |
| `custom` | `openai-responses` | `POST <base>/responses` | `Authorization: Bearer` |
| `azure-foundry` | — | `POST <base>/openai/v1/chat/completions?api-version=…` | `api-key` (model→deployment) |

The base URL is normalized like omni-agent-desktop: trailing slashes are
stripped, and if the URL has no path after the host, `/v1` is appended
(so `http://localhost:5000` becomes `http://localhost:5000/v1`).

### Settings & precedence

Every setting can come from a CLI flag, an environment variable, the config
file, or a built-in default. Higher in this list wins:

**CLI flag > env var > config file > default**

| Setting | Flag | Env var | Config key | Default |
|---|---|---|---|---|
| Provider | `-p`, `--provider` | `PROMPTSMITH_PROVIDER` | `provider` | `custom` |
| API shape | `-s`, `--api-shape` | `PROMPTSMITH_API_SHAPE` | `api_shape` | `openai-compatible` |
| Base URL | `-u`, `--base-url` | `PROMPTSMITH_BASE_URL` | `base_url` | `http://localhost:5000/v1` |
| Model | `-m`, `--model` | `PROMPTSMITH_MODEL` | `model` | `gpt-5.5` |
| API key | `-k`, `--api-key` | `PROMPTSMITH_API_KEY` | `api_key` | falls back to `~/.config/omnillm/api-key` |
| Temperature | `-t`, `--temperature` | — | — | `0.3` |

### Config file

`~/.config/promptsmith/config.json` (created by `install.sh`). Full schema:

```json
{
  "provider": "custom",
  "api_shape": "openai-compatible",
  "base_url": "http://localhost:5000/v1",
  "model": "gpt-5.5",
  "api_key": "",
  "azure_api_version": "2024-02-01",
  "azure_deployments": [
    { "model": "gpt-4o", "deployment": "my-gpt4o-deployment" }
  ]
}
```

Leave `api_key` empty to use an env var or the OmniLLM key file instead. Keeping
the key out of the config file and in `PROMPTSMITH_API_KEY` is recommended.

### Provider recipes

**OmniLLM (default, local proxy)** — nothing to configure. promptsmith points at
`http://localhost:5000/v1` (openai-compatible) and reads the key from
`~/.config/omnillm/api-key` automatically. Just run `promptsmith "..."`.

OmniLLM also speaks the Anthropic and Responses shapes for the right models:

```bash
promptsmith -s anthropic-messages -m claude-opus-4.8 "polish this"
promptsmith -s openai-responses   -m gpt-5.5          "polish this"
```

**OpenAI** (openai-compatible)

```json
{ "provider": "custom", "api_shape": "openai-compatible",
  "base_url": "https://api.openai.com/v1", "model": "gpt-4o" }
```
```bash
export PROMPTSMITH_API_KEY="sk-..."
```

**Anthropic** (native Messages API, `anthropic-messages` shape)

```json
{ "provider": "custom", "api_shape": "anthropic-messages",
  "base_url": "https://api.anthropic.com", "model": "claude-3-5-sonnet-latest" }
```
```bash
export PROMPTSMITH_API_KEY="sk-ant-..."
```

**Azure AI Foundry** (`api-key` header, logical model → deployment mapping)

```json
{
  "provider": "azure-foundry",
  "base_url": "https://my-resource.openai.azure.com",
  "azure_api_version": "2024-02-01",
  "model": "gpt-4o",
  "azure_deployments": [{ "model": "gpt-4o", "deployment": "my-gpt4o-deployment" }]
}
```
```bash
export PROMPTSMITH_API_KEY="<azure-key>"
```

**Groq / OpenRouter / together / etc.** — any OpenAI-compatible endpoint:

```json
{ "base_url": "https://api.groq.com/openai/v1", "model": "llama-3.3-70b-versatile" }
{ "base_url": "https://openrouter.ai/api/v1",   "model": "anthropic/claude-3.5-sonnet" }
```

**Local vLLM / llama.cpp / Ollama** — OpenAI-compatible servers usually ignore
the key, but a non-empty placeholder is still required:

```bash
promptsmith -u http://localhost:8000/v1  -m meta-llama/Llama-3.1-8B-Instruct -k EMPTY "..."
promptsmith -u http://localhost:11434/v1 -m llama3.1 -k ollama "..."
```

### One-off overrides

Flags override everything for a single run without touching your config:

```bash
promptsmith -p custom -s anthropic-messages -u https://api.anthropic.com \
  -m claude-3-5-sonnet-latest -k "$ANTHROPIC_API_KEY" "improve this"
```

### Verify your setup

```bash
promptsmith --list-models     # confirms base URL + key reach the provider
```

Works for the `openai-compatible` shape and `azure-foundry` (both expose
`GET /models`). If it prints models, you're wired up correctly; a connection or
`HTTP 401` error points to a wrong `base_url` or `api_key`.

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
