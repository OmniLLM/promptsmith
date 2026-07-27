package main

// Interactive polish shell: a single-session, conversational REPL for iterating
// on one prompt. The first line you type is polished from scratch; every line
// after that is treated as a refine/change request applied to the current
// polished prompt (the same engine as --iterate), so a whole conversation keeps
// sharpening one prompt. Meta-commands start with ":".

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// runShell drives the interactive session. It reuses polish() for the first
// turn and runIterate() for every refinement, so no new provider plumbing is
// needed. cfg/key/temp are captured once from the resolved CLI config.
func runShell(cfg config, key string, temp float64) {
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 1<<20)

	printShellBanner(cfg)

	var current string // the working polished prompt, "" until the first turn

	for {
		if current == "" {
			fmt.Print(shellPrompt("polish"))
		} else {
			fmt.Print(shellPrompt("refine"))
		}
		if !in.Scan() {
			fmt.Println()
			break // EOF (Ctrl-D) or read error
		}
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}

		// Meta-commands.
		if strings.HasPrefix(line, ":") {
			if quit := handleShellCommand(line, &current); quit {
				break
			}
			continue
		}

		// Normal turn: first message polishes, later messages refine.
		if current == "" {
			fmt.Println(dim("  polishing…"))
			out := polish(cfg, cfg.BaseURL, key, cfg.Model, true, temp, line)
			current = extractPrompt(out)
		} else {
			fmt.Println(dim("  refining…"))
			out := runIterate(cfg, key, temp, current, line, true)
			current = extractPrompt(out)
		}
		fmt.Println()
		fmt.Println(renderMarkdown("## Current prompt\n```text\n" + current + "\n```"))
		fmt.Println()
	}
	if err := in.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "pps: input error:", err)
	}
	fmt.Println(dim("bye."))
}

// extractPrompt pulls just the clean prompt body out of a model response. Even
// in --raw mode, models often wrap the prompt in a heading ("## Optimized
// Prompt"), fence it, and append a "Notes on what I changed" explanation. We
// keep only the prompt so the shell iterates on the prompt itself, not on the
// model's commentary. Preference order:
//  1. the contents of the first fenced code block, if any;
//  2. otherwise, the text with a leading heading dropped and any trailing
//     explanation section ("Notes…", "---", "If you tell me…") cut off.
func extractPrompt(out string) string {
	s := strings.TrimSpace(out)
	if body, ok := firstFencedBlock(s); ok {
		return strings.TrimSpace(body)
	}
	lines := strings.Split(s, "\n")

	// Drop a leading markdown heading like "## Optimized Prompt" (and a blank
	// line or "---" rule right after it).
	for len(lines) > 0 {
		t := strings.TrimSpace(lines[0])
		if t == "" || t == "---" || t == "***" || strings.HasPrefix(t, "#") {
			lines = lines[1:]
			continue
		}
		break
	}

	// Cut everything from the first explanation marker onward.
	for i, l := range lines {
		t := strings.TrimSpace(l)
		low := strings.ToLower(strings.Trim(t, "*_ "))
		if t == "---" || t == "***" ||
			strings.HasPrefix(low, "notes on what") ||
			strings.HasPrefix(low, "notes:") ||
			strings.HasPrefix(low, "what i changed") ||
			strings.HasPrefix(low, "changes made") ||
			strings.HasPrefix(low, "if you tell me") ||
			strings.HasPrefix(low, "if you can tell me") ||
			strings.HasPrefix(low, "let me know") {
			lines = lines[:i]
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// firstFencedBlock returns the contents of the first ```-fenced block in s.
func firstFencedBlock(s string) (string, bool) {
	lines := strings.Split(s, "\n")
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			start = i
			break
		}
	}
	if start < 0 {
		return "", false
	}
	for j := start + 1; j < len(lines); j++ {
		if strings.HasPrefix(strings.TrimSpace(lines[j]), "```") {
			return strings.Join(lines[start+1:j], "\n"), true
		}
	}
	return "", false // unterminated fence; fall back to heuristic stripping
}

// handleShellCommand runs a ":" meta-command. Returns true to end the session.
func handleShellCommand(line string, current *string) bool {
	cmd, arg, _ := strings.Cut(strings.TrimPrefix(line, ":"), " ")
	arg = strings.TrimSpace(arg)
	switch cmd {
	case "q", "quit", "exit":
		return true
	case "h", "help", "?":
		printShellHelp()
	case "show":
		if *current == "" {
			fmt.Println(dim("  (no prompt yet — type one to polish it)"))
		} else {
			fmt.Println(renderMarkdown("```text\n" + *current + "\n```"))
		}
	case "reset", "new":
		*current = ""
		fmt.Println(dim("  session cleared — next line starts a fresh polish"))
	case "raw":
		if *current == "" {
			fmt.Println(dim("  (nothing to copy yet)"))
		} else {
			// Unstyled, flush-left, no frame — for clean copy/redirect.
			fmt.Println(*current)
		}
	case "save":
		if *current == "" {
			fmt.Println(dim("  (nothing to save yet)"))
		} else if arg == "" {
			fmt.Println(yellow("  usage: :save <file>"))
		} else if err := os.WriteFile(arg, []byte(*current+"\n"), 0o644); err != nil {
			fmt.Println(yellow("  cannot write " + arg + ": " + err.Error()))
		} else {
			fmt.Println(green("  wrote " + arg))
		}
	default:
		fmt.Println(yellow("  unknown command :" + cmd + "  (try :help)"))
	}
	return false
}

func shellPrompt(kind string) string {
	arrow := "›"
	label := kind + " " + arrow + " "
	if kind == "polish" {
		return cyan(label)
	}
	return green(label)
}

func printShellBanner(cfg config) {
	fmt.Println(bold("promptsmith interactive shell") +
		dim("  ("+cfg.Provider+" · "+cfg.Model+")"))
	fmt.Println(dim("Type a prompt to polish it, then keep talking to refine it. :help for commands, :quit to exit."))
	fmt.Println()
}

func printShellHelp() {
	rows := [][]string{
		{":show", "print the current polished prompt (framed)"},
		{":raw", "print it unstyled/flush-left for clean copy or redirect"},
		{":save <file>", "write the current prompt to a file"},
		{":reset", "clear the session and start a fresh polish"},
		{":help", "show this help"},
		{":quit", "exit (or Ctrl-D)"},
	}
	fmt.Println(renderTable([]string{"Command", "What it does"}, rows))
	fmt.Println(dim("Anything not starting with ':' is your message: the first polishes, the rest refine."))
}
