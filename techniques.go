package main

// Technique catalog: the 17 prompt-engineering techniques from
// promptingguide.ai. The full reference guide for each is embedded from
// techniques/*.md so the binary can print a guide or inject it into the system prompt when the
// user pins a technique with -T/--technique.

import (
	"embed"
	"fmt"
	"os"
	"sort"
	"strings"
)

//go:embed techniques/*.md
var techniqueFS embed.FS

// selectedTechniques holds the techniques pinned via -T/--technique.
var selectedTechniques []technique

type technique struct {
	Name    string   // canonical name, matches techniques/<name>.md
	Aliases []string // short forms accepted on the CLI
	Summary string   // one-line what it is
	UseWhen string   // symptom that calls for it
}

var techniques = []technique{
	{"zero-shot", []string{"zs"},
		"Clear instruction, no examples. Cheapest option — always try first.",
		"Task is common and the model likely already knows it."},
	{"few-shot", []string{"fs", "fewshot"},
		"Show 2-8 input/output demonstrations to teach format and label space.",
		"Output needs a specific format, label set, or style."},
	{"chain-of-thought", []string{"cot"},
		"Elicit intermediate reasoning steps before the answer.",
		"Multi-step reasoning, math, logic, or symbolic manipulation."},
	{"self-consistency", []string{"sc"},
		"Sample N reasoning paths at higher temperature, take the majority answer.",
		"Answers are inconsistent run to run on a task with one right answer."},
	{"generated-knowledge", []string{"gk", "knowledge"},
		"Have the model generate relevant facts first, then answer using them.",
		"Task needs world knowledge the model has but doesn't surface unprompted."},
	{"tree-of-thoughts", []string{"tot"},
		"Explore multiple reasoning branches, evaluate, backtrack, and prune.",
		"Complex planning/search where a single linear chain gets stuck."},
	{"rag", []string{"retrieval"},
		"Retrieve external documents and ground the answer in them with citations.",
		"Needs current, private, or verifiable facts the model can't know."},
	{"react", []string{"reason-act"},
		"Interleave Thought → Action → Observation loops against real tools.",
		"Needs tools/search plus reasoning; pure CoT would hallucinate facts."},
	{"art", []string{"auto-reasoning"},
		"Automatic Reasoning and Tool-use: pick task exemplars and tool calls from a library.",
		"Multi-step tool workflows you want generalized rather than hand-written."},
	{"pal", []string{"program-aided"},
		"Offload computation to generated code executed by an interpreter.",
		"Arithmetic, dates, data manipulation — anything a program does exactly."},
	{"reflexion", []string{"reflect"},
		"Agent critiques its own failed attempt and retries with the lesson in context.",
		"Iterative agent tasks where the first attempt often fails."},
	{"meta-prompting", []string{"meta"},
		"Specify structure and syntax abstractly rather than through examples.",
		"You want a token-thrifty skeleton and format contract, not demonstrations."},
	{"ape", []string{"auto-prompt"},
		"Automatic Prompt Engineer: generate candidate instructions and score them.",
		"You can evaluate outputs automatically and want the instruction optimized."},
	{"active-prompt", []string{"active"},
		"Pick which examples to annotate by measuring model uncertainty.",
		"You have a labeling budget and want the highest-value few-shot exemplars."},
	{"directional-stimulus", []string{"ds", "stimulus"},
		"Inject hints/keywords that steer the output toward desired content.",
		"Summaries or generations that keep missing specific required content."},
	{"multimodal-cot", []string{"mcot"},
		"Chain-of-thought over text plus images, reasoning across both modalities.",
		"Task involves images/diagrams as well as text."},
	{"graphprompt", []string{"graph"},
		"Cast the task over graph-structured data with a unified prompt template.",
		"Inputs are graphs/relations — nodes, edges, structured entity links."},
}

// techniqueCatalog renders the whole catalog as a compact menu the evaluator
// can pick from. Keeping it inline (rather than embedding all 17 full guides)
// keeps the eval call cheap while still constraining recommendations to real,
// named techniques instead of invented ones.
func techniqueCatalog() string {
	var sb strings.Builder
	for _, t := range techniques {
		fmt.Fprintf(&sb, "- %s — %s Use when: %s\n", t.Name, t.Summary, t.UseWhen)
	}
	return sb.String()
}

func findTechnique(q string) (technique, bool) {
	q = strings.ToLower(strings.TrimSpace(q))
	for _, t := range techniques {
		if t.Name == q {
			return t, true
		}
		for _, a := range t.Aliases {
			if a == q {
				return t, true
			}
		}
	}
	return technique{}, false
}

func (t technique) guide() string {
	b, err := techniqueFS.ReadFile("techniques/" + t.Name + ".md")
	if err != nil {
		return ""
	}
	return string(b)
}

func techniqueNames() string {
	names := make([]string, 0, len(techniques))
	for _, t := range techniques {
		names = append(names, t.Name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// printTechniques lists the catalog for --list-techniques.
func printTechniques() {
	fmt.Println(bold("promptsmith techniques") + dim("  (source: https://www.promptingguide.ai/techniques)"))
	fmt.Println()
	headers := []string{"Technique", "Aliases", "Summary", "Use when"}
	var rows [][]string
	for _, t := range techniques {
		rows = append(rows, []string{
			t.Name,
			strings.Join(t.Aliases, ", "),
			t.Summary,
			t.UseWhen,
		})
	}
	fmt.Println(renderTable(headers, rows))
	fmt.Println()
	fmt.Println(dim("Pin one or more with:  ") + "pps -T cot,few-shot \"your prompt\"")
	fmt.Println(dim("Read a full guide with: ") + "pps --show-technique react")
}

// showTechnique prints the full reference guide for one technique.
func showTechnique(name string) {
	t, ok := findTechnique(name)
	if !ok {
		fail("unknown technique %q. Known: %s", name, techniqueNames())
	}
	g := t.guide()
	if g == "" {
		fail("no embedded guide for %q", t.Name)
	}
	fmt.Println(strings.TrimRight(g, "\n"))
}

// resolveTechniques parses a comma-separated -T value into techniques,
// failing loudly on an unknown name.
func resolveTechniques(spec string) []technique {
	var out []technique
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		t, ok := findTechnique(part)
		if !ok {
			fmt.Fprintf(os.Stderr, "pps: unknown technique %q\n", part)
			fmt.Fprintf(os.Stderr, "known techniques: %s\n", techniqueNames())
			fmt.Fprintf(os.Stderr, "run `pps --list-techniques` for descriptions\n")
			os.Exit(1)
		}
		out = append(out, t)
	}
	return out
}

// techniqueDirective builds the system-prompt addendum that pins the model to
// the requested techniques and hands it the full reference guide for each.
func techniqueDirective(sel []technique) string {
	if len(sel) == 0 {
		return ""
	}
	names := make([]string, 0, len(sel))
	for _, t := range sel {
		names = append(names, t.Name)
	}
	var sb strings.Builder
	sb.WriteString("\n\n### TECHNIQUE OVERRIDE (user-specified)\n")
	sb.WriteString("The user has explicitly requested these technique(s): ")
	sb.WriteString(strings.Join(names, ", "))
	sb.WriteString(".\nSkip step 3 of the workflow (technique selection) — apply exactly these, ")
	sb.WriteString("stacked, even if you would have chosen differently. In the ")
	sb.WriteString("\"Technique(s) applied\" section, name each one and explain concretely how ")
	sb.WriteString("it shaped the rewritten prompt. If a requested technique is a poor fit for ")
	sb.WriteString("the task, still apply it, but add a one-line caveat noting the mismatch and ")
	sb.WriteString("what you would have used instead.\n")
	sb.WriteString("\nReference guides for the requested technique(s):\n")
	for _, t := range sel {
		g := t.guide()
		if g == "" {
			continue
		}
		sb.WriteString("\n<technique name=\"" + t.Name + "\">\n")
		sb.WriteString(strings.TrimSpace(g))
		sb.WriteString("\n</technique>\n")
	}
	return sb.String()
}
