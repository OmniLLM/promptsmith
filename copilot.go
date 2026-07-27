// GitHub Copilot provider — OAuth device flow + short-lived Copilot API token.
//
// Auth model (mirrors OmniLLM's internal/providers/copilot):
//  1. Device-code OAuth against github.com yields a long-lived GitHub token.
//  2. That token is exchanged at api.github.com/copilot_internal/v2/token for a
//     short-lived Copilot API token (~30 min) plus the account's API host —
//     enterprise seats get a non-public host, so always adopt endpoints.api.
//  3. Requests go to <host>/chat/completions with Bearer + the editor headers
//     Copilot expects.
//
// Credentials live in ~/.config/promptsmith/copilot.json (mode 0600).
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ghBaseURL    = "https://github.com"
	ghAPIBaseURL = "https://api.github.com"
	// Public client ID of the GitHub Copilot editor OAuth app (not a secret).
	ghClientID     = "Iv1.b507a08c87ecfe98"
	ghScopes       = "read:user"
	copilotAPIHost = "https://api.githubcopilot.com"
	editorVersion  = "vscode/1.83.1"
	pluginVersion  = "copilot-chat/0.26.7"
	copilotUA      = "GitHubCopilotChat/0.26.7"
	ghAPIVersion   = "2025-04-01"
)

// copilotCreds is the on-disk credential cache.
type copilotCreds struct {
	GitHubToken  string `json:"github_token"`  // long-lived OAuth token
	CopilotToken string `json:"copilot_token"` // short-lived API token
	ExpiresAt    int64  `json:"expires_at"`    // unix expiry of CopilotToken
	APIHost      string `json:"api_host"`      // endpoints.api from the exchange
	Login        string `json:"login,omitempty"`
}

func copilotCredsPath() string {
	return filepath.Join(home(), ".config", "promptsmith", "copilot.json")
}

func loadCopilotCreds() copilotCreds {
	var c copilotCreds
	if b, err := os.ReadFile(copilotCredsPath()); err == nil {
		_ = json.Unmarshal(b, &c)
	}
	return c
}

func saveCopilotCreds(c copilotCreds) {
	p := copilotCredsPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		fail("cannot create %s — %v", filepath.Dir(p), err)
	}
	b, _ := json.MarshalIndent(c, "", "  ")
	if err := os.WriteFile(p, append(b, '\n'), 0o600); err != nil {
		fail("cannot write %s — %v", p, err)
	}
}

func ghPostJSON(url string, payload any, out any) error {
	buf, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d from %s — %s", resp.StatusCode, url, truncate(string(body), 300))
	}
	return json.Unmarshal(body, out)
}

// copilotLogin runs the device-code flow interactively and caches the result.
func copilotLogin() {
	var dc struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURI string `json:"verification_uri"`
		ExpiresIn       int    `json:"expires_in"`
		Interval        int    `json:"interval"`
	}
	if err := ghPostJSON(ghBaseURL+"/login/device/code",
		map[string]string{"client_id": ghClientID, "scope": ghScopes}, &dc); err != nil {
		fail("device code request failed — %v", err)
	}

	fmt.Fprintf(os.Stderr, "\n  Open %s\n  Enter code: %s\n\nWaiting for authorization…\n",
		dc.VerificationURI, dc.UserCode)

	interval := time.Duration(dc.Interval+1) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(dc.ExpiresIn) * time.Second)

	var ghToken string
	for time.Now().Before(deadline) {
		time.Sleep(interval)
		var tr struct {
			AccessToken string `json:"access_token"`
			Error       string `json:"error"`
		}
		err := ghPostJSON(ghBaseURL+"/login/oauth/access_token", map[string]string{
			"client_id":   ghClientID,
			"device_code": dc.DeviceCode,
			"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
		}, &tr)
		if err != nil {
			continue // transient; keep polling
		}
		if tr.AccessToken != "" {
			ghToken = tr.AccessToken
			break
		}
		switch tr.Error {
		case "expired_token", "access_denied":
			fail("authorization failed: %s", tr.Error)
		case "slow_down":
			interval += 5 * time.Second
		}
	}
	if ghToken == "" {
		fail("device code expired before authorization")
	}

	creds := copilotCreds{GitHubToken: ghToken}
	creds = refreshCopilotToken(creds)
	creds.Login = githubLogin(ghToken)
	saveCopilotCreds(creds)

	who := creds.Login
	if who == "" {
		who = "your account"
	}
	fmt.Fprintf(os.Stderr, "pps: authorized as %s — credentials saved to %s\n",
		who, copilotCredsPath())
}

func githubLogin(ghToken string) string {
	req, _ := http.NewRequest("GET", ghAPIBaseURL+"/user", nil)
	req.Header.Set("Authorization", "token "+ghToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", copilotUA)
	resp, err := httpClient().Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var u struct {
		Login string `json:"login"`
	}
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &u)
	return u.Login
}

// refreshCopilotToken exchanges the GitHub OAuth token for a Copilot API token
// and adopts the account-specific API host advertised by the exchange.
func refreshCopilotToken(c copilotCreds) copilotCreds {
	if c.GitHubToken == "" {
		fail("not logged in to GitHub Copilot — run: pps --copilot-login")
	}
	req, _ := http.NewRequest("GET", ghAPIBaseURL+"/copilot_internal/v2/token", nil)
	req.Header.Set("Authorization", "token "+c.GitHubToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Editor-Version", editorVersion)
	req.Header.Set("Editor-Plugin-Version", pluginVersion)
	req.Header.Set("User-Agent", copilotUA)
	req.Header.Set("X-Github-Api-Version", ghAPIVersion)

	resp, err := httpClient().Do(req)
	if err != nil {
		fail("copilot token exchange failed — %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fail("copilot token exchange failed with HTTP %d — %s\n"+
			"(does this account have an active Copilot seat? try: pps --copilot-login)",
			resp.StatusCode, truncate(string(body), 300))
	}
	var tr struct {
		Token     string `json:"token"`
		ExpiresAt int64  `json:"expires_at"`
		Endpoints struct {
			API string `json:"api"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(body, &tr); err != nil {
		fail("unexpected copilot token response: %s", truncate(string(body), 300))
	}
	c.CopilotToken = tr.Token
	c.ExpiresAt = tr.ExpiresAt
	// Enterprise seats serve from their own host; personal seats get the public one.
	if api := strings.TrimSpace(tr.Endpoints.API); api != "" {
		c.APIHost = api
	} else if c.APIHost == "" {
		c.APIHost = copilotAPIHost
	}
	return c
}

// copilotSession returns a valid Copilot API token and its base host, refreshing
// (and re-persisting) the short-lived token when it is within 5 min of expiry.
func copilotSession() (token, host string) {
	c := loadCopilotCreds()
	if c.GitHubToken == "" {
		// Allow a pre-existing GitHub token from the environment.
		if env := strings.TrimSpace(os.Getenv("PROMPTSMITH_GITHUB_TOKEN")); env != "" {
			c.GitHubToken = env
		} else {
			fail("not logged in to GitHub Copilot — run: pps --copilot-login")
		}
	}
	if c.CopilotToken == "" || c.ExpiresAt == 0 || time.Now().Unix() > c.ExpiresAt-300 {
		c = refreshCopilotToken(c)
		saveCopilotCreds(c)
	}
	host = c.APIHost
	if host == "" {
		host = copilotAPIHost
	}
	return c.CopilotToken, strings.TrimRight(host, "/")
}

func copilotRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	h := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}

func copilotHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization":                       "Bearer " + token,
		"Accept":                              "application/json",
		"copilot-integration-id":              "vscode-chat",
		"Editor-Version":                      editorVersion,
		"Editor-Plugin-Version":               pluginVersion,
		"User-Agent":                          copilotUA,
		"OpenAI-Intent":                       "conversation-panel",
		"X-Github-Api-Version":                ghAPIVersion,
		"X-Request-Id":                        copilotRequestID(),
		"X-Initiator":                         "user",
		"X-Vscode-User-Agent-Library-Version": "electron-fetch",
	}
}

// polishCopilot sends a system+user pair to the Copilot API. Newer models
// (GPT-5-class) are responses-only and reject /chat/completions with
// unsupported_api_for_model — transparently fall back to /responses.
func polishCopilot(model, system, user string, temp float64) string {
	return chatCopilot(model, temp, system, []chatMsg{{Role: "user", Content: user}})
}

// chatCopilot sends a full conversation to Copilot's chat-completions endpoint,
// falling back to /responses for responses-only models.
func chatCopilot(model string, temp float64, system string, history []chatMsg) string {
	token, host := copilotSession()
	url := host + "/chat/completions"
	payload := map[string]any{
		"model":       model,
		"messages":    oaMessages(system, history),
		"temperature": temp,
	}
	body, status, raw := tryPOST(url, copilotHeaders(token), payload)
	if status == 200 {
		return parseChatCompletions(body, url)
	}
	if isUnsupportedChatCompletions(status, raw) {
		// Responses fallback only carries the latest user turn's text; flatten
		// the conversation into one instruction+message pair.
		return copilotResponses(token, host, model, system, flattenHistory(history))
	}
	fail("HTTP %d from %s — %s", status, url, truncate(raw, 500))
	return ""
}

// flattenHistory renders a conversation as plain text for endpoints that take a
// single user message rather than a turn list.
func flattenHistory(history []chatMsg) string {
	var sb strings.Builder
	for _, m := range history {
		label := "User"
		if m.Role == "assistant" {
			label = "Assistant"
		}
		sb.WriteString(label + ": " + m.Content + "\n\n")
	}
	return strings.TrimSpace(sb.String())
}

// isUnsupportedChatCompletions detects Copilot's responses-only model rejection.
func isUnsupportedChatCompletions(status int, body string) bool {
	if status != http.StatusBadRequest {
		return false
	}
	lower := strings.ToLower(body)
	return strings.Contains(lower, "unsupported_api_for_model") ||
		(strings.Contains(lower, "not accessible") && strings.Contains(lower, "chat/completions"))
}

// copilotResponses calls the /responses endpoint for responses-only models.
func copilotResponses(token, host, model, system, user string) string {
	url := host + "/responses"
	payload := map[string]any{
		"model":        model,
		"instructions": system,
		"input": []map[string]any{{
			"role":    "user",
			"content": []map[string]string{{"type": "input_text", "text": user}},
		}},
		"store": false,
	}
	body := doPOST(url, copilotHeaders(token), payload)
	var b struct {
		OutputText string `json:"output_text"`
		Output     []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.Unmarshal(body, &b); err != nil {
		fail("unexpected copilot responses payload: %s", truncate(string(body), 500))
	}
	if strings.TrimSpace(b.OutputText) != "" {
		return b.OutputText
	}
	var sb strings.Builder
	for _, item := range b.Output {
		if item.Type != "message" {
			continue // skip reasoning items
		}
		for _, c := range item.Content {
			sb.WriteString(c.Text)
		}
	}
	if strings.TrimSpace(sb.String()) == "" {
		fail("empty copilot responses payload: %s", truncate(string(body), 500))
	}
	return sb.String()
}

// listCopilotModels prints the models the account's Copilot seat exposes.
func listCopilotModels() {
	token, host := copilotSession()
	url := host + "/models"
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range copilotHeaders(token) {
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
	ids := make([]string, len(ml.Data))
	for i, m := range ml.Data {
		ids[i] = m.ID
	}
	printModelList(ids, "github-copilot")
}

// copilotStatus reports the cached credential state without making a chat call.
func copilotStatus() {
	c := loadCopilotCreds()
	if c.GitHubToken == "" {
		fmt.Println("github-copilot: not logged in (run: pps --copilot-login)")
		return
	}
	who := c.Login
	if who == "" {
		who = "(unknown login)"
	}
	fmt.Printf("github-copilot: logged in as %s\n", who)
	host := c.APIHost
	if host == "" {
		host = copilotAPIHost
	}
	fmt.Printf("  api host    : %s\n", host)
	fmt.Printf("  creds file  : %s\n", copilotCredsPath())
	if c.ExpiresAt > 0 {
		left := time.Until(time.Unix(c.ExpiresAt, 0)).Round(time.Second)
		if left > 0 {
			fmt.Printf("  api token   : valid for %s\n", left)
		} else {
			fmt.Printf("  api token   : expired (auto-refreshes on next call)\n")
		}
	}
}

// copilotLogout removes the cached credentials.
func copilotLogout() {
	p := copilotCredsPath()
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		fail("cannot remove %s — %v", p, err)
	}
	fmt.Fprintf(os.Stderr, "pps: removed %s\n", p)
}
