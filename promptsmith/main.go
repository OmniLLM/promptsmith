// promptsmith — a CLI prompt optimizer.
//
// Takes a raw prompt / task description and returns a polished, technique-backed
// version using proven prompt-engineering techniques (zero-shot, few-shot, CoT,
// self-consistency, ReAct, meta-prompting, ToT, RAG, PAL, Reflexion, and more —
// distilled from promptingguide.ai).
//
// Talks to any OpenAI-compatible endpoint. Defaults to a local OmniLLM proxy at
// http://localhost:5000/v1.
//
// Usage:
//
//	promptsmith "write a tweet about cats"
//	echo "summarize this article" | promptsmith
//	promptsmith -m claude-opus-4.8 "classify sentiment"
//	promptsmith --raw "just give me the rewritten prompt, no explanation"
//	promptsmith --list-models
//
// Config (precedence: CLI flag > env var > config file > default):
//
//	PROMPTSMITH_BASE_URL   (default http://localhost:5000/v1)
//	PROMPTSMITH_API_KEY    (default: read ~/.config/omnillm/api-key)
//	PROMPTSMITH_MODEL      (default gpt-5.5)
//	Config file: ~/.config/promptsmith/config.json
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultBaseURL = "http://localhost:5000/v1"
	defaultModel   = "gpt-5.5"
)

const systemPrompt = `You are promptsmith, an expert prompt-engineering coach. Your job is to take a
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
  (e.g. ` + "`Sentiment:`, `A:`" + `, a JSON schema) — format alone lifts performance.
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
` + "```" + `
<the rewritten prompt, ready to paste>
` + "```" + `

## Knobs to tune
<shots / temperature / tools / retrieval source, as relevant>`

const rawSuffix = "\n\nIMPORTANT: The user wants ONLY the polished prompt itself. Output the " +
	"rewritten prompt as plain text with no headings, no explanation, no code " +
	"fences, no commentary. Just the prompt, ready to paste."

type config struct {
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type modelList struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func home() string {
	h, _ := os.UserHomeDir()
	return h
}

func loadConfig() config {
	var c config
	p := filepath.Join(home(), ".config", "promptsmith", "config.json")
	b, err := os.ReadFile(p)
	if err == nil {
		_ = json.Unmarshal(b, &c)
	}
	return c
}

func resolveAPIKey(cfg config) string {
	if v := os.Getenv("PROMPTSMITH_API_KEY"); v != "" {
		return strings.TrimSpace(v)
	}
	if cfg.APIKey != "" {
		return strings.TrimSpace(cfg.APIKey)
	}
	p := filepath.Join(home(), ".config", "omnillm", "api-key")
	if b, err := os.ReadFile(p); err == nil {
		return strings.TrimSpace(string(b))
	}
	return ""
}

func pick(flagVal, env, cfgVal, def string) string {
	if flagVal != "" {
		return flagVal
	}
	if v := os.Getenv(env); v != "" {
		return v
	}
	if cfgVal != "" {
		return cfgVal
	}
	return def
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "promptsmith: "+format+"\n", a...)
	os.Exit(1)
}

func httpClient() *http.Client {
	return &http.Client{Timeout: 300 * time.Second}
}

func listModels(baseURL, apiKey string) {
	req, _ := http.NewRequest("GET", baseURL+"/models", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := httpClient().Do(req)
	if err != nil {
		fail("cannot reach %s/models — %v", baseURL, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fail("HTTP %d from %s/models — %s", resp.StatusCode, baseURL, truncate(string(body), 500))
	}
	var ml modelList
	if err := json.Unmarshal(body, &ml); err != nil {
		fail("unexpected /models response: %s", truncate(string(body), 300))
	}
	for _, m := range ml.Data {
		fmt.Println(m.ID)
	}
}

func polish(baseURL, apiKey, model, promptText string, raw bool, temperature float64) string {
	system := systemPrompt
	if raw {
		system += rawSuffix
	}
	user := "Optimize the following prompt. Here is the raw prompt/task " +
		"description:\n\n" + strings.TrimSpace(promptText)

	reqBody := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: temperature,
	}
	buf, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient().Do(req)
	if err != nil {
		fail("cannot reach %s — %v. Is OmniLLM running?", baseURL, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fail("HTTP %d from %s — %s", resp.StatusCode, baseURL, truncate(string(body), 500))
	}
	var cr chatResponse
	if err := json.Unmarshal(body, &cr); err != nil || len(cr.Choices) == 0 {
		fail("unexpected response: %s", truncate(string(body), 500))
	}
	return cr.Choices[0].Message.Content
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func usage() {
	fmt.Fprint(os.Stderr, `promptsmith — optimize prompts using proven prompt-engineering techniques via an LLM.

Usage:
  promptsmith [flags] "raw prompt to optimize"
  echo "raw prompt" | promptsmith [flags]

Flags:
  -m, --model string    Model (default gpt-5.5)
  -u, --base-url string OpenAI-compatible base URL (default http://localhost:5000/v1)
  -k, --api-key string  API key (default: env or ~/.config/omnillm/api-key)
  -t, --temperature f   Sampling temperature (default 0.3)
      --raw             Output only the polished prompt, no explanation
      --list-models     List available models and exit
  -h, --help            Show this help
`)
}

func main() {
	var (
		model, baseURL, apiKey        string
		temperature                   float64
		raw, listModelsFlag, helpFlag bool
	)
	fs := flag.NewFlagSet("promptsmith", flag.ContinueOnError)
	fs.Usage = usage
	fs.StringVar(&model, "m", "", "model")
	fs.StringVar(&model, "model", "", "model")
	fs.StringVar(&baseURL, "u", "", "base url")
	fs.StringVar(&baseURL, "base-url", "", "base url")
	fs.StringVar(&apiKey, "k", "", "api key")
	fs.StringVar(&apiKey, "api-key", "", "api key")
	fs.Float64Var(&temperature, "t", 0.3, "temperature")
	fs.Float64Var(&temperature, "temperature", 0.3, "temperature")
	fs.BoolVar(&raw, "raw", false, "raw output")
	fs.BoolVar(&listModelsFlag, "list-models", false, "list models")
	fs.BoolVar(&helpFlag, "h", false, "help")
	fs.BoolVar(&helpFlag, "help", false, "help")

	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if helpFlag {
		usage()
		return
	}

	cfg := loadConfig()
	baseURL = strings.TrimRight(pick(baseURL, "PROMPTSMITH_BASE_URL", cfg.BaseURL, defaultBaseURL), "/")
	model = pick(model, "PROMPTSMITH_MODEL", cfg.Model, defaultModel)
	if apiKey == "" {
		apiKey = resolveAPIKey(cfg)
	}

	if listModelsFlag {
		listModels(baseURL, apiKey)
		return
	}

	var promptText string
	if args := fs.Args(); len(args) > 0 {
		promptText = strings.Join(args, " ")
	} else if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		b, _ := io.ReadAll(os.Stdin)
		promptText = string(b)
	} else {
		usage()
		os.Exit(1)
	}

	if strings.TrimSpace(promptText) == "" {
		fail("empty prompt.")
	}

	out := polish(baseURL, apiKey, model, promptText, raw, temperature)
	fmt.Println(strings.TrimRight(out, "\n"))
}
