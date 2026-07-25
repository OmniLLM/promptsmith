#!/usr/bin/env python3
"""
promptsmith — a CLI prompt optimizer.

Takes a raw prompt / task description and returns a polished, technique-backed
version using proven prompt-engineering techniques (zero-shot, few-shot, CoT,
self-consistency, ReAct, meta-prompting, ToT, RAG, PAL, Reflexion, and more —
distilled from promptingguide.ai).

Talks to any OpenAI-compatible endpoint. Defaults to a local OmniLLM proxy at
http://localhost:5000/v1.

Usage:
  promptsmith "write a tweet about cats"
  echo "summarize this article" | promptsmith
  promptsmith -m claude-opus-4.8 "classify sentiment"
  promptsmith --raw "just give me the rewritten prompt, no explanation"
  promptsmith --list-models

Config (precedence: CLI flag > env var > config file > default):
  PROMPTSMITH_BASE_URL   (default http://localhost:5000/v1)
  PROMPTSMITH_API_KEY    (default: read ~/.config/omnillm/api-key)
  PROMPTSMITH_MODEL      (default gpt-5.5)
  Config file: ~/.config/promptsmith/config.json
"""

import argparse
import json
import os
import sys
import urllib.request
import urllib.error
from pathlib import Path

DEFAULT_BASE_URL = "http://localhost:5000/v1"
DEFAULT_MODEL = "gpt-5.5"
CONFIG_PATH = Path.home() / ".config" / "promptsmith" / "config.json"
OMNILLM_KEY_PATH = Path.home() / ".config" / "omnillm" / "api-key"

SYSTEM_PROMPT = r"""You are promptsmith, an expert prompt-engineering coach. Your job is to take a
user's raw prompt (or task description) and return a polished, technique-backed
version, plus a short rationale for why those techniques fit.

You draw on the 17 techniques documented at promptingguide.ai:
zero-shot, few-shot, chain-of-thought, self-consistency, generated-knowledge,
tree-of-thoughts, RAG, ReAct, ART, PAL, Reflexion, meta-prompting, APE,
active-prompt, directional-stimulus, multimodal-CoT, graphprompt.

WORKFLOW (follow every time):
1. Understand the task. Identify task type (classification / extraction /
   reasoning / math / coding / creative / factual-QA / summarization /
   decision-making / agentic), and whether it needs reasoning steps, external
   facts/tools, examples, or just a clearer instruction.
2. Diagnose weaknesses in the current prompt: vague instruction, no output
   format, no examples, no reasoning scaffold, no grounding, etc.
3. Select 1-3 technique(s) by symptom:
   - simple, model likely knows it        -> Zero-shot (clear instruction + format anchor)
   - needs specific format/label space    -> Few-shot (2-8 demonstrations)
   - multi-step reasoning / math / logic  -> Chain-of-Thought ("think step by step")
   - inconsistent answers                 -> Self-Consistency (sample N, majority vote)
   - missing world knowledge              -> Generated Knowledge
   - complex, needs exploration           -> Tree of Thoughts
   - needs current/private grounding      -> RAG
   - needs tools + reasoning interleaved  -> ReAct / ART
   - should run/verify code               -> PAL
   - agent learning from mistakes         -> Reflexion
   - want skeleton/structure, token-thrift-> Meta-Prompting
4. Rewrite the prompt applying the chosen technique(s). Keep the user's intent;
   add structure, examples, format anchors, reasoning triggers, or grounding.

CORE POLISHING PRINCIPLES (apply on top of any technique):
- Be explicit about the task AND the output format. Anchor the format
  (e.g. `Sentiment:`, `A:`, a JSON schema) — format alone lifts performance.
- Put instructions first, then context, then the input. Use clear delimiters
  (###, triple backticks, XML tags) to separate sections.
- Trigger reasoning BEFORE the answer for anything non-trivial. Models that
  answer first then justify tend to rationalize a wrong answer.
- Prefer showing over telling — one good example beats a paragraph of rules.
- Match examples to the true label distribution; keep a consistent format.
- Stack techniques when useful (Few-shot + CoT, ReAct + CoT + Self-Consistency).
- Stop when it's good enough — don't add tokens a simple task doesn't need.
  Zero-shot first; escalate only on failure.

OUTPUT FORMAT (unless the user asks for raw output only):

## Diagnosis
<what's weak in the current prompt — 1-4 bullets>

## Technique(s) applied
<technique — one line why, for each>

## Polished prompt
```
<the rewritten prompt, ready to paste>
```

## Knobs to tune
<shots / temperature / tools / retrieval source, as relevant>
"""

RAW_SUFFIX = (
    "\n\nIMPORTANT: The user wants ONLY the polished prompt itself. Output the "
    "rewritten prompt as plain text with no headings, no explanation, no code "
    "fences, no commentary. Just the prompt, ready to paste."
)


def load_config():
    cfg = {}
    if CONFIG_PATH.exists():
        try:
            cfg = json.loads(CONFIG_PATH.read_text())
        except Exception:
            pass
    return cfg


def resolve_api_key(cfg):
    key = os.environ.get("PROMPTSMITH_API_KEY") or cfg.get("api_key")
    if key:
        return key.strip()
    if OMNILLM_KEY_PATH.exists():
        return OMNILLM_KEY_PATH.read_text().strip()
    return ""


def resolve(name, flag_val, env, cfg_key, default):
    if flag_val:
        return flag_val
    if os.environ.get(env):
        return os.environ[env]
    if cfg_key in cfg_cache:
        return cfg_cache[cfg_key]
    return default


cfg_cache = {}


def http_post(url, api_key, payload, timeout=180):
    data = json.dumps(payload).encode()
    req = urllib.request.Request(
        url,
        data=data,
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {api_key}",
        },
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read().decode())


def http_get(url, api_key, timeout=30):
    req = urllib.request.Request(
        url, headers={"Authorization": f"Bearer {api_key}"}, method="GET"
    )
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read().decode())


def list_models(base_url, api_key):
    try:
        d = http_get(f"{base_url}/models", api_key)
        for m in d.get("data", []):
            print(m.get("id", ""))
    except urllib.error.URLError as e:
        sys.exit(f"promptsmith: cannot reach {base_url}/models — {e}")


def polish(base_url, api_key, model, prompt_text, raw=False, temperature=0.3, stream=False):
    system = SYSTEM_PROMPT + (RAW_SUFFIX if raw else "")
    user = (
        "Optimize the following prompt. Here is the raw prompt/task "
        "description:\n\n" + prompt_text.strip()
    )
    payload = {
        "model": model,
        "messages": [
            {"role": "system", "content": system},
            {"role": "user", "content": user},
        ],
        "temperature": temperature,
    }
    try:
        resp = http_post(f"{base_url}/chat/completions", api_key, payload)
    except urllib.error.HTTPError as e:
        body = e.read().decode(errors="replace")
        sys.exit(f"promptsmith: HTTP {e.code} from {base_url} — {body[:500]}")
    except urllib.error.URLError as e:
        sys.exit(f"promptsmith: cannot reach {base_url} — {e}. Is OmniLLM running?")
    try:
        return resp["choices"][0]["message"]["content"]
    except (KeyError, IndexError):
        sys.exit(f"promptsmith: unexpected response: {json.dumps(resp)[:500]}")


def main():
    global cfg_cache
    p = argparse.ArgumentParser(
        prog="promptsmith",
        description="Optimize prompts using proven prompt-engineering techniques via an LLM.",
    )
    p.add_argument("prompt", nargs="*", help="Raw prompt to optimize (or pipe via stdin).")
    p.add_argument("-m", "--model", help=f"Model (default {DEFAULT_MODEL}).")
    p.add_argument("-u", "--base-url", help=f"OpenAI-compatible base URL (default {DEFAULT_BASE_URL}).")
    p.add_argument("-k", "--api-key", help="API key (default: env or ~/.config/omnillm/api-key).")
    p.add_argument("-t", "--temperature", type=float, default=0.3, help="Sampling temperature (default 0.3).")
    p.add_argument("--raw", action="store_true", help="Output only the polished prompt, no explanation.")
    p.add_argument("--list-models", action="store_true", help="List available models and exit.")
    args = p.parse_args()

    cfg_cache = load_config()
    base_url = (args.base_url or os.environ.get("PROMPTSMITH_BASE_URL")
                or cfg_cache.get("base_url") or DEFAULT_BASE_URL).rstrip("/")
    api_key = args.api_key or resolve_api_key(cfg_cache)
    model = (args.model or os.environ.get("PROMPTSMITH_MODEL")
             or cfg_cache.get("model") or DEFAULT_MODEL)

    if args.list_models:
        list_models(base_url, api_key)
        return

    if args.prompt:
        prompt_text = " ".join(args.prompt)
    elif not sys.stdin.isatty():
        prompt_text = sys.stdin.read()
    else:
        p.print_help()
        sys.exit(1)

    if not prompt_text.strip():
        sys.exit("promptsmith: empty prompt.")

    out = polish(base_url, api_key, model, prompt_text, raw=args.raw, temperature=args.temperature)
    print(out.rstrip())


if __name__ == "__main__":
    main()
