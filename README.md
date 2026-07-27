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
cd promptsmith
./install.sh          # builds the Go binary into ~/.local/bin + scaffolds config
```

Or build manually:

```bash
cd promptsmith
go build -o pps .
```

Written in Go — no runtime dependencies, single static binary. Requires Go >= 1.21 to build.

## Usage

```bash
pps                                              # interactive polish shell (see below)
pps "write a tweet about cats"
echo "summarize this article" | pps
pps -m claude-opus-4.8 "classify sentiment of reviews"
pps --raw "just give me the rewritten prompt"   # polished prompt only, no explanation
pps --list-models                                # list available models
```

## Interactive shell

Run `pps` with no prompt (or `pps --chat` / `pps -i`) on a terminal to open a
conversational session. Type a prompt to polish it, then keep talking to refine
that *same* prompt turn by turn — the first message is polished from scratch and
every following message is applied as a change request to the current prompt.

```
promptsmith interactive shell  (custom · claude-opus-4.8)
polish › classify sentiment of reviews
refine › make it output JSON
refine › add a confidence score
refine › :save prompt.md
```

Meta-commands (anything starting with `:`):

| Command | What it does |
|---|---|
| `:show` | print the current polished prompt (framed) |
| `:raw` | print it unstyled/flush-left for clean copy or redirect |
| `:save <file>` | write the current prompt to a file |
| `:reset` | clear the session and start a fresh polish |
| `:help` | list commands |
| `:quit` | exit (Ctrl-D also works) |

Piped or argument invocations stay one-shot, so scripts are unaffected.

Read from and write to files:

```bash
pps -f prompt.md                       # optimize a prompt stored in a file
pps -f prompt.md --raw -o prompt.v2.md # write the result out (also printed to stdout)
```

Default output is structured:

```
## Diagnosis          what's weak in your prompt
## Technique(s) applied  which techniques and why
## Techniques considered  what it evaluated and rejected, and why
## Polished prompt    the rewritten prompt, ready to paste
## Knobs to tune      shots / temperature / tools / retrieval source
```

Use `--raw` to get only the rewritten prompt (great for piping/scripting).

On an interactive terminal the structured output is rendered with color, section
rules, and aligned tables (`--list-techniques`, `--list-modes`, `--list-models`
and any Markdown tables the optimizer emits). When stdout is piped, redirected,
or `NO_COLOR` is set, output degrades to plain Markdown / ASCII tables so
scripting and file writes (`-o`) stay clean. The polished-prompt block is always
emitted flush-left and unstyled, so selecting it copies the prompt verbatim.

## Techniques

By default promptsmith picks the technique(s) itself by diagnosing your prompt.
You can also browse the catalog and pin specific ones.

```bash
pps --list-techniques            # all 17 techniques + when to use each
pps --show-technique react       # full reference guide for one technique
pps -T cot "solve this puzzle"   # force chain-of-thought
pps -T few-shot,cot "classify"   # stack several
```

`-T` / `--technique` takes a comma-separated list. Each name accepts short
aliases (`cot`, `fs`, `tot`, `sc`, `zs`, `gk`, `ds`, `mcot`, …). When you pin
techniques, promptsmith skips its own selection step, injects the full reference
guide for each into the system prompt, and applies exactly what you asked for —
adding a one-line caveat if a technique is a poor fit for the task.

| Technique | Use when |
|---|---|
| `zero-shot` (`zs`) | Task is common and the model likely already knows it |
| `few-shot` (`fs`) | Output needs a specific format, label set, or style |
| `chain-of-thought` (`cot`) | Multi-step reasoning, math, logic |
| `self-consistency` (`sc`) | Answers are inconsistent run to run |
| `generated-knowledge` (`gk`) | Needs world knowledge it doesn't surface unprompted |
| `tree-of-thoughts` (`tot`) | Complex planning/search; a linear chain gets stuck |
| `rag` | Needs current, private, or verifiable facts |
| `react` | Needs tools/search interleaved with reasoning |
| `art` | Multi-step tool workflows you want generalized |
| `pal` | Arithmetic, dates, data manipulation — anything code does exactly |
| `reflexion` | Iterative agent tasks where the first attempt often fails |
| `meta-prompting` (`meta`) | Want a token-thrifty skeleton and format contract |
| `ape` | Outputs are auto-evaluable and you want the instruction optimized |
| `active-prompt` | Limited labeling budget; pick the highest-value exemplars |
| `directional-stimulus` (`ds`) | Generations keep missing required content |
| `multimodal-cot` (`mcot`) | Task involves images/diagrams as well as text |
| `graphprompt` | Inputs are graphs/relations — nodes, edges, entity links |

The reference guides are embedded in the binary at build time from
`techniques/`, so `--show-technique`
works with no network and no repo checkout.

## Modes and styles

Optimizing a **system prompt** (a persistent role definition) is a different job
from optimizing a **user prompt** (a single request). promptsmith splits them,
with three styles each. Run `pps --list-modes` to see them.

```bash
pps --mode system --style analytical "you are a code reviewer"
pps --mode user   --style planning   "help me launch a newsletter"
```

| `--mode` | `--style` | What it produces |
|---|---|---|
| `system` (default) | `general` (default) | Role, goal, capabilities, rules, workflow, output requirements |
| `system` | `analytical` | Adds a reasoning procedure, evidence discipline, edge cases, self-check |
| `system` | `output-format` | Locks down an exact machine-checkable output contract with an example |
| `user` | `basic` (default) | Sharpens a vague request: intent, context, format, audience |
| `user` | `planning` | Decomposes a fuzzy goal into an ordered, executable step plan |
| `user` | `professional` | Injects domain terminology, standards, and the expected rigor |

Every mode enforces two guardrails: the model must **rewrite** the prompt rather
than execute it, and it must **preserve `{{variable}}` placeholders** verbatim
instead of filling them with concrete values (which would destroy a reusable
template).

## Target model class

Classic prompt-engineering advice was written for instruction-following models
and **partially inverts for reasoning models**. promptsmith branches on this:

```bash
pps --target instruct  "solve seating-arrangement logic puzzles"
pps --target reasoning "solve seating-arrangement logic puzzles"
```

| | instruction-following (GPT-4-class, Sonnet-class, local) | reasoning (o-series, GPT-5-class, extended thinking) |
|---|---|---|
| mental model | junior coworker — spell it out | senior coworker — give the goal, trust the route |
| chain-of-thought | add explicit "think step by step" triggers | **do not add them** — the model reasons internally and triggers can hurt |
| examples | few-shot helps | try zero-shot first; mismatched examples hurt more |
| length | detailed scaffolding is fine | keep it brief and direct |
| emphasis | prescribe the steps | prescribe the destination, the constraints, and how to verify "done" |

Without `--target`, promptsmith infers the class from the prompt and states the
assumption in *Knobs to tune*. The same input produces visibly different output:
the `instruct` version emits a numbered procedure plus a reasoning trigger; the
`reasoning` version emits a goal, constraints, an output contract, and a
self-verification clause with **no** chain-of-thought instruction at all.

## What the optimizer knows

Beyond the 17 techniques, the doctrine baked into promptsmith's own system
prompt covers:

- **Four elements of a prompt** — instruction, context, input data, output
  indicator — with missing elements called out in the diagnosis.
- **Say what to do, not what not to do.** `DO NOT ASK FOR INTERESTS` reliably
  backfires; every prohibition is converted into a replacement behavior plus a
  fallback for when the model can't comply.
- **Precision over cleverness.** Vague qualifiers (short, detailed,
  professional) become countable limits, a named audience, and a concrete style.
- **Prompt budget.** Detail helps only while it stays relevant; lines that don't
  change behavior get cut. This is what stops an optimizer from bloating.
- **Structure and placement.** Identity → Instructions → Examples → Context,
  with bulky context last. For **long context**, instructions go *both above and
  below* the document — empirically better than either alone. Later
  instructions win conflicts, so contradictions are flagged rather than guessed.
- **Literal instruction following.** Modern models obey absolute rules
  absolutely, including in cases you didn't consider — so every hard rule gets
  an escape hatch.
- **Safety scaffolding, scaled to risk.** Untrusted input (`{{user_input}}`,
  retrieved docs, emails) gets structural isolation plus an explicit
  "this is data, not instructions" clause; factual tasks get grounding, a
  permitted "I don't know", and a ban on invented specifics; failure paths get a
  concrete fallback string. A haiku prompt gets none of this.
- **Exemplar hygiene.** Balanced label distribution, randomized order,
  identical formatting — skewed or grouped examples bias the output.
- **Sampling config with real numbers.** Balanced 0.2/0.95/30, creative
  0.9/0.99/40, low-variance 0.1/0.90/20, and temperature 0 for extraction,
  classification, JSON, and chain-of-thought. It also knows that token limits
  *truncate* rather than summarize, and that repetition loops can come from
  temperature being too low **or** too high.

## Iterating on a prompt

Once you have a prompt you like, refine it against a specific change request
instead of re-optimizing from scratch:

```bash
pps -f prompt.md --iterate "add a JSON output contract"
pps -f prompt.md --iterate "make it refuse out-of-scope questions" --raw -o prompt.v2.md
```

This makes a surgical edit: it folds the request in as a new instruction or
constraint, keeps existing constraints and `{{placeholders}}` intact, and does
not rewrite sections the request didn't touch. Default output shows
`## Changes made` plus the updated prompt; `--raw` gives just the prompt.

The distinction that makes this work: `--iterate "no back-and-forth"` adds a
*no-clarifying-questions constraint to the prompt* — it does not make the model
reply "ok, I won't ask you questions".

## Scoring a prompt

```bash
pps -f prompt.md --eval                    # human-readable report
pps -f prompt.md --eval --json             # raw JSON, for scripting
pps -f prompt.md --eval --min-score 70     # CI gate: exit 2 if below
```

Scores the prompt 0-100 on five design dimensions, proposes concrete repairs,
**and recommends which prompt-engineering techniques to apply**:

```
## Score: 16/100

  Goal Clarity                 ███████░░░░░░░░░░░░░  38
  Instruction Completeness     ██░░░░░░░░░░░░░░░░░░  12
  Structural Executability     ███░░░░░░░░░░░░░░░░░  18
  Ambiguity Control            ██░░░░░░░░░░░░░░░░░░  10
  Robustness                   ░░░░░░░░░░░░░░░░░░░░   4

## Recommended techniques

- **rag** (HIGH)
    why: the prompt asks for a buy/sell verdict but names no data source or
         as-of date, so the model will invent figures.
    how: require retrieval of filings/quotes as of a stated date, cite each
         number with its source, and refuse a verdict if data is unavailable.

- **meta-prompting** (HIGH)
    why: "how is it doing" defines no evaluation dimensions or output shape.
    how: add a fixed output contract — thesis, fundamentals, valuation, risks,
         scenarios, information gaps, conditional verdict.

Apply them:  pps -T rag,meta-prompting -f <prompt-file>

## Techniques considered and rejected
- few-shot — the gap is missing inputs and evidence, not output style.
- chain-of-thought — naming the dimensions beats asking for visible reasoning.
```

Recommendations are constrained to the 17 techniques in the catalog (the same
names `-T` accepts), so the report ends with a copy-pasteable command that
actually applies them. The rejected list is deliberate: showing what was
*considered and dismissed*, with reasons, is what makes the advice accountable
rather than a black-box verdict.

The `--json` form includes a `patchPlan` where each `oldText` is an exact
substring of the input, so edits can be applied programmatically:

```json
{
  "score": { "overall": 16, "dimensions": [ … ] },
  "improvements": ["…"],
  "patchPlan": [
    { "op": "replace", "oldText": "<exact fragment>", "newText": "…", "instruction": "issue + fix" }
  ],
  "techniqueRecommendations": [
    { "technique": "rag", "priority": "high", "why": "…", "how": "…" }
  ],
  "techniquesRejected": [
    { "technique": "few-shot", "reason": "…" }
  ],
  "summary": "One-sentence verdict"
}
```

`--eval` clamps temperature to ≤0.1 so scores are as stable as the model allows.

### CI gating

`--min-score N` turns `--eval` into a check you can put in a pipeline:

```bash
pps -f prompts/*.md --eval --min-score 70
```

Exit codes: `0` = at or above the bar, `2` = scored below it (the report still
prints), `1` = the tool itself failed (network, unparseable response). The
distinct exit 2 lets CI tell "prompt needs work" apart from "the check broke".

## A/B comparing two prompts

`--eval` judges a prompt on paper. `--compare` judges it by *running* it: both
prompts are executed against the same test input (concurrently), then a judge
model scores the two outputs.

```bash
pps -f v2.md --compare v1.md --test "review this login function: ..."
pps -f v2.md --compare v1.md --test "..." --show-outputs   # also print both outputs
pps -f v2.md --compare v1.md --test "..." --json           # machine-readable verdict
```

A is the `--compare` file (baseline), B is the main input (candidate).

```
## Winner: B (candidate)   (confidence: high)

## Scores
                                  A        B    delta
  Goal Achievement                7        9       +2
  Output Quality                  6        9       +3
  Constraint Compliance           8       10       +2
  Prompt Effectiveness            6        9       +3

## Generalization
  overfit risk:   low
  generalizes:    yes
  recommendation: adopt
```

The **over-fitting guard** is the point. A prompt can win on one test input
while being worse in general — because it hard-codes an answer, narrows scope
to that case, or bakes in assumptions that only hold there. The judge is
required to separate "better on this sample" from "generalizably better" and to
say when the sample is too thin to tell. `recommendation` is one of `adopt`,
`keep-original`, or `needs-more-tests`.

## Making a prompt reusable

`--templatize` finds the parts of a concrete prompt that would change between
uses and replaces them with `{{variables}}`:

```bash
pps -f prompt.md --templatize --vars-out vars.json
pps -f prompt.md --templatize --max-vars 3
```

```
## Variables extracted (4)

  {{code_to_review}}  [subject]
      default: submitted Django code
      why:     Reuse the same review prompt for different codebases.
  ...

## Templatized prompt
You are a senior Python code reviewer. Review the {{code_to_review}} for
{{review_focus}}. ...
```

Each variable's `originalText` must be an exact substring of the input, so
substitution is mechanical rather than a re-generation — the prompt's wording
is never silently rewritten. Anything that can't be matched verbatim is
reported as skipped instead of being applied blindly.

Then fill the template back in — **locally, with no API call**:

```bash
pps -f tpl.md --render --var topic=Kubernetes --var tone=technical
pps -f tpl.md --render --vars vars.json
pps -f tpl.md --render --vars vars.json --strict   # exit 1 if any {{var}} is left
```

`--render` is instant and free. `--var` overrides `--vars`. Without `--strict`,
unfilled placeholders are reported on stderr but the render still succeeds.

templatize → render is a lossless round trip: rendering a freshly templatized
prompt with its own `vars.json` reproduces the original text exactly.

### A full loop

```bash
pps --mode system --style analytical --raw "you are a code reviewer" -o v1.md
pps -f v1.md --eval
pps -f v1.md --iterate "require a severity label per finding" --raw -o v2.md
pps -f v2.md --eval                                    # score on paper
pps -f v2.md --compare v1.md --test "<a real case>"    # score in practice
pps -f v2.md --templatize --vars-out vars.json         # make it reusable
```

## Configuring the LLM provider

pps supports the **same providers and wire shapes as
[omni-agent-desktop](https://github.com/OmniLLM/omni-agent-desktop)**, so a
config that works there works here. Two provider types, and the custom provider
speaks one of three API shapes:

| Provider (`-p`) | API shape (`-s`) | Request | Auth header |
|---|---|---|---|
| `custom` | `openai-compatible` (default) | `POST <base>/chat/completions` | `Authorization: Bearer` |
| `custom` | `anthropic-messages` | `POST <base>/messages` | `x-api-key` + `anthropic-version` |
| `custom` | `openai-responses` | `POST <base>/responses` | `Authorization: Bearer` |
| `azure-foundry` | — | `POST <base>/openai/v1/chat/completions?api-version=…` | `api-key` (model→deployment) |
| `github-copilot` | — | `POST <copilot-host>/chat/completions` | OAuth device flow → Bearer (no API key) |

### GitHub Copilot (OAuth)

`-p github-copilot` uses your **GitHub Copilot seat** instead of an API key.
Auth is the standard device flow, same as the VS Code Copilot extension:

```bash
pps --copilot-login     # open github.com/login/device, enter the code
pps --copilot-status    # who you are + token validity
pps -p github-copilot --list-models
pps -p github-copilot -m gpt-4.1 "review this SQL migration"
pps --copilot-logout    # delete cached credentials
```

How it works:

1. Device-code OAuth against `github.com` yields a **long-lived GitHub token**.
2. That token is exchanged at `api.github.com/copilot_internal/v2/token` for a
   **short-lived Copilot API token** (~30 min) plus the account's API host.
   Enterprise seats get a private host (`api.enterprise.githubcopilot.com`);
   promptsmith always adopts whatever `endpoints.api` reports, so enterprise
   seats route correctly.
3. The short-lived token auto-refreshes when it is within 5 minutes of expiry —
   you log in once.

Credentials are cached in `~/.config/promptsmith/copilot.json` (mode `0600`).
`-k/--api-key`, `-u/--base-url` and `-s/--api-shape` are ignored for this
provider — the host comes from the token exchange. Default model is `gpt-4.1`
(Copilot has its own catalog; `--list-models` shows what your seat exposes).
An existing GitHub OAuth token can be supplied via `PROMPTSMITH_GITHUB_TOKEN`
instead of running `--copilot-login`.

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
| Model | `-m`, `--model` | `PROMPTSMITH_MODEL` | `model` | `gpt-5.5` (`gpt-4.1` for `github-copilot`) |
| API key | `-k`, `--api-key` | `PROMPTSMITH_API_KEY` | `api_key` | falls back to `~/.config/omnillm/api-key` |
| GitHub token | — | `PROMPTSMITH_GITHUB_TOKEN` | `copilot.json` | from `--copilot-login` |
| Temperature | `-t`, `--temperature` | — | — | `0.3` |
| Technique(s) | `-T`, `--technique` | — | — | auto-selected |
| Mode | `--mode` | — | — | `system` |
| Style | `--style` | — | — | mode's first style |
| Target model class | `--target` | — | — | inferred |

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
`~/.config/omnillm/api-key` automatically. Just run `pps "..."`.

OmniLLM also speaks the Anthropic and Responses shapes for the right models:

```bash
pps -s anthropic-messages -m claude-opus-4.8 "polish this"
pps -s openai-responses   -m gpt-5.5          "polish this"
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
pps -u http://localhost:8000/v1  -m meta-llama/Llama-3.1-8B-Instruct -k EMPTY "..."
pps -u http://localhost:11434/v1 -m llama3.1 -k ollama "..."
```

### One-off overrides

Flags override everything for a single run without touching your config:

```bash
pps -p custom -s anthropic-messages -u https://api.anthropic.com \
  -m claude-3-5-sonnet-latest -k "$ANTHROPIC_API_KEY" "improve this"
```

### Verify your setup

```bash
pps --list-models     # confirms base URL + key reach the provider
```

Works for the `openai-compatible` shape and `azure-foundry` (both expose
`GET /models`). If it prints models, you're wired up correctly; a connection or
`HTTP 401` error points to a wrong `base_url` or `api_key`.

## Hermes skill

This repo also ships the original **`prompt-engineering`** Hermes Agent skill
(the methodology the CLI is built on) as `SKILL.md`, backed by a full reference
guide per technique under `techniques/`. Install it with:

```bash
ln -s "$PWD" ~/.hermes/skills/prompt-engineering
```

## Layout

```
pps/
  main.go               single-file Go CLI (stdlib only)
  techniques.go         technique catalog + embedded reference guides
  modes.go              optimization modes/styles, iterate + eval templates
  compare.go            A/B compare, templatize, local variable render
  copilot.go            GitHub Copilot OAuth device flow + token refresh
  shell.go              interactive polish shell (conversational refine loop)
  render.go             TTY-aware ANSI styling, tables, Markdown-to-terminal
  techniques/           the 17 reference guides (embedded at build time)
  go.mod
  install.sh            installer (go build + config scaffold)
  SKILL.md              the Hermes skill (methodology + technique selector)
  PROMPTSMITH_SYSTEM_PROMPT.md   promptsmith's own agent system prompt
```

## promptsmith's own system prompt

The identity promptsmith runs on is documented as a portable, paste-anywhere
system prompt in
[`PROMPTSMITH_SYSTEM_PROMPT.md`](PROMPTSMITH_SYSTEM_PROMPT.md)
— use it to give any LLM or agent promptsmith behaviour without the CLI.

To see the *live* prompt the CLI actually sends (base doctrine plus the active
mode/style/target/technique directives), run:

```bash
pps --show-system
pps --show-system --mode user --style planning --target reasoning -T cot
```

## Credits

The mode/style split, the iterate flow, the `{{variable}}` preservation rule,
the evaluation rubric, the over-fitting guards in A/B comparison, and variable
extraction are adapted from
[linshenkx/prompt-optimizer](https://github.com/linshenkx/prompt-optimizer).
The technique catalog comes from [promptingguide.ai](https://www.promptingguide.ai/techniques).

The doctrine in the optimizer's system prompt additionally draws on
[promptingguide.ai](https://www.promptingguide.ai/) (elements of a prompt,
general tips, adversarial/factuality/bias risk pages), Google's *Prompt
Engineering* whitepaper by Lee Boonstra (sampling configuration, the repetition
loop, instructions over constraints, few-shot class mixing), Anthropic's
prompting best practices (XML delimiting, grounding and abstention, long-context
quote extraction), and OpenAI's reasoning-model guidance and GPT-4.1 cookbook
(the reasoning-vs-instruct split, literal instruction following, long-context
instruction placement).
