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
			current = stripFence(strings.TrimSpace(out))
		} else {
			fmt.Println(dim("  refining…"))
			out := runIterate(cfg, key, temp, current, line, true)
			current = stripFence(strings.TrimSpace(out))
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
