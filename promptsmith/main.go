// promptsmith — a CLI prompt optimizer.
//
// Takes a raw prompt / task description and returns a polished, technique-backed
// version using proven prompt-engineering techniques (zero-shot, few-shot, CoT,
// self-consistency, ReAct, meta-prompting, ToT, RAG, PAL, Reflexion, and more —
// distilled from promptingguide.ai).
//
// Providers mirror omni-agent-desktop:
//   - custom provider with one of three API shapes:
//     openai-compatible   POST <base>/chat/completions   (Bearer auth)
//     anthropic-messages  POST <base>/messages           (x-api-key)
//     openai-responses    POST <base>/responses          (Bearer auth)
//   - azure-foundry         POST <base>/openai/v1/chat/completions?api-version=…
//     (api-key header, model→deployment)
//
// Defaults to a local OmniLLM proxy at http://localhost:5000/v1 (openai-compatible).
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
	defaultShape   = "openai-compatible"
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

// azureDeployment maps a logical model name to a concrete Azure deployment.
type azureDeployment struct {
	Model      string `json:"model"`
	Deployment string `json:"deployment"`
}

// config mirrors the provider fields of omni-agent-desktop's ProviderConfig.
type config struct {
	Provider         string            `json:"provider"`  // "custom" | "azure-foundry"
	APIShape         string            `json:"api_shape"` // for custom
	BaseURL          string            `json:"base_url"`  // endpoint
	Model            string            `json:"model"`
	APIKey           string            `json:"api_key"`
	AzureAPIVersion  string            `json:"azure_api_version"` // azure-foundry
	AzureDeployments []azureDeployment `json:"azure_deployments"` // azure-foundry
}

func home() string {
	h, _ := os.UserHomeDir()
	return h
}

func loadConfig() config {
	var c config
	p := filepath.Join(home(), ".config", "promptsmith", "config.json")
	if b, err := os.ReadFile(p); err == nil {
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
	// OmniLLM default key file (matches the default endpoint).
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

// normalizeEndpoint mirrors omni-agent-desktop: strip trailing slashes; if the
// URL has no path segment after the host, append /v1.
func normalizeEndpoint(endpoint string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if trimmed == "" {
		return trimmed
	}
	afterScheme := trimmed
	if i := strings.Index(trimmed, "://"); i >= 0 {
		afterScheme = trimmed[i+3:]
	}
	if strings.Contains(afterScheme, "/") {
		return trimmed
	}
	return trimmed + "/v1"
}

func doPOST(url string, headers map[string]string, payload any) []byte {
	buf, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(buf))
	req.Header.Set("content-type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := httpClient().Do(req)
	if err != nil {
		fail("cannot reach %s — %v. Is the provider running?", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fail("HTTP %d from %s — %s", resp.StatusCode, url, truncate(string(body), 500))
	}
	return body
}

// --- provider request/response shapes ---

func polishOpenAIChat(base, key, model, system, user string, temp float64) string {
	url := normalizeEndpoint(base) + "/chat/completions"
	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": temp,
	}
	body := doPOST(url, map[string]string{"authorization": "Bearer " + key}, payload)
	return parseChatCompletions(body, url)
}

func polishAzure(cfg config, key, system, user string, temp float64) string {
	deployment := resolveDeployment(cfg)
	apiVersion := cfg.AzureAPIVersion
	if apiVersion == "" {
		apiVersion = "2024-02-01"
	}
	base := strings.TrimRight(cfg.BaseURL, "/")
	url := fmt.Sprintf("%s/openai/v1/chat/completions?api-version=%s", base, apiVersion)
	payload := map[string]any{
		"model": deployment,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": temp,
		"store":       false,
	}
	body := doPOST(url, map[string]string{"api-key": key}, payload)
	return parseChatCompletions(body, url)
}

func polishAnthropic(base, key, model, system, user string) string {
	url := normalizeEndpoint(base) + "/messages"
	payload := map[string]any{
		"model":  model,
		"system": system,
		"messages": []map[string]string{
			{"role": "user", "content": user},
		},
		"max_tokens": 4096,
	}
	headers := map[string]string{
		"x-api-key":         key,
		"anthropic-version": "2023-06-01",
	}
	body := doPOST(url, headers, payload)
	var b struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &b); err != nil {
		fail("unexpected anthropic response: %s", truncate(string(body), 500))
	}
	var sb strings.Builder
	for _, blk := range b.Content {
		if blk.Type == "text" {
			sb.WriteString(blk.Text)
		}
	}
	return sb.String()
}

func polishResponses(base, key, model, system, user string) string {
	url := normalizeEndpoint(base) + "/responses"
	payload := map[string]any{
		"model":        model,
		"instructions": system,
		"input":        []map[string]string{{"role": "user", "content": user}},
	}
	body := doPOST(url, map[string]string{"authorization": "Bearer " + key}, payload)
	var b struct {
		OutputText string `json:"output_text"`
		Output     []struct {
			Type    string `json:"type"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.Unmarshal(body, &b); err != nil {
		fail("unexpected responses payload: %s", truncate(string(body), 500))
	}
	if b.OutputText != "" {
		return b.OutputText
	}
	for _, item := range b.Output {
		for _, blk := range item.Content {
			if blk.Text != "" {
				return blk.Text
			}
		}
	}
	fail("empty responses payload: %s", truncate(string(body), 500))
	return ""
}

func parseChatCompletions(body []byte, url string) string {
	var cr struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &cr); err != nil || len(cr.Choices) == 0 {
		fail("unexpected response from %s: %s", url, truncate(string(body), 500))
	}
	return cr.Choices[0].Message.Content
}

func resolveDeployment(cfg config) string {
	model := strings.TrimSpace(cfg.Model)
	for _, m := range cfg.AzureDeployments {
		if m.Model == model {
			return m.Deployment
		}
	}
	return model
}

// listModels hits GET /models where the shape supports it (openai-compatible /
// azure). Anthropic/responses shapes don't expose a compatible listing.
func listModels(cfg config, base, key string) {
	var url string
	headers := map[string]string{}
	switch {
	case cfg.Provider == "azure-foundry":
		apiVersion := cfg.AzureAPIVersion
		if apiVersion == "" {
			apiVersion = "2024-02-01"
		}
		url = fmt.Sprintf("%s/openai/v1/models?api-version=%s", strings.TrimRight(cfg.BaseURL, "/"), apiVersion)
		headers["api-key"] = key
	default:
		url = normalizeEndpoint(base) + "/models"
		headers["authorization"] = "Bearer " + key
	}
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := httpClient().Do(req)
	if err != nil {
		fail("cannot reach %s — %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fail("HTTP %d from %s — %s", resp.StatusCode, url, truncate(string(body), 500))
	}
	var ml struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &ml); err != nil {
		fail("unexpected /models response: %s", truncate(string(body), 300))
	}
	for _, m := range ml.Data {
		fmt.Println(m.ID)
	}
}

func polish(cfg config, base, key, model string, raw bool, temp float64, promptText string) string {
	system := systemPrompt
	if raw {
		system += rawSuffix
	}
	user := "Optimize the following prompt. Here is the raw prompt/task " +
		"description:\n\n" + strings.TrimSpace(promptText)

	if cfg.Provider == "azure-foundry" {
		return polishAzure(cfg, key, system, user, temp)
	}
	switch cfg.APIShape {
	case "anthropic-messages":
		return polishAnthropic(base, key, model, system, user)
	case "openai-responses":
		return polishResponses(base, key, model, system, user)
	default: // openai-compatible
		return polishOpenAIChat(base, key, model, system, user, temp)
	}
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
  -m, --model string    Model / Azure logical model name (default gpt-5.5)
  -u, --base-url string Endpoint / base URL (default http://localhost:5000/v1)
  -k, --api-key string  API key (default: env or ~/.config/omnillm/api-key)
  -p, --provider string Provider: custom | azure-foundry (default custom)
  -s, --api-shape string  Custom provider wire shape:
                          openai-compatible | anthropic-messages | openai-responses
                          (default openai-compatible)
  -t, --temperature f   Sampling temperature (default 0.3)
      --raw             Output only the polished prompt, no explanation
      --list-models     List available models and exit
  -h, --help            Show this help

Config file (~/.config/promptsmith/config.json) can set provider, api_shape,
base_url, model, api_key, azure_api_version, azure_deployments.
`)
}

func main() {
	var (
		model, baseURL, apiKey, provider, apiShape string
		temperature                                float64
		raw, listModelsFlag, helpFlag              bool
	)
	fs := flag.NewFlagSet("promptsmith", flag.ContinueOnError)
	fs.Usage = usage
	fs.StringVar(&model, "m", "", "model")
	fs.StringVar(&model, "model", "", "model")
	fs.StringVar(&baseURL, "u", "", "base url")
	fs.StringVar(&baseURL, "base-url", "", "base url")
	fs.StringVar(&apiKey, "k", "", "api key")
	fs.StringVar(&apiKey, "api-key", "", "api key")
	fs.StringVar(&provider, "p", "", "provider")
	fs.StringVar(&provider, "provider", "", "provider")
	fs.StringVar(&apiShape, "s", "", "api shape")
	fs.StringVar(&apiShape, "api-shape", "", "api shape")
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
	cfg.BaseURL = strings.TrimRight(pick(baseURL, "PROMPTSMITH_BASE_URL", cfg.BaseURL, defaultBaseURL), "/")
	cfg.Model = pick(model, "PROMPTSMITH_MODEL", cfg.Model, defaultModel)
	cfg.Provider = pick(provider, "PROMPTSMITH_PROVIDER", cfg.Provider, "custom")
	cfg.APIShape = pick(apiShape, "PROMPTSMITH_API_SHAPE", cfg.APIShape, defaultShape)
	if apiKey == "" {
		apiKey = resolveAPIKey(cfg)
	}

	if listModelsFlag {
		listModels(cfg, cfg.BaseURL, apiKey)
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

	out := polish(cfg, cfg.BaseURL, apiKey, cfg.Model, raw, temperature, promptText)
	fmt.Println(strings.TrimRight(out, "\n"))
}
