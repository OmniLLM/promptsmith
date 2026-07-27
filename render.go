package main

// Terminal presentation layer: TTY-aware ANSI styling, aligned ASCII/box tables,
// and a lightweight Markdown-to-terminal renderer. Everything degrades to plain
// text when stdout is not a terminal (piped/redirected) or NO_COLOR is set, so
// scripting and file output stay clean. No third-party dependencies.

import (
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// stdoutIsTTY reports whether stdout is an interactive terminal. Same character-
// device check the stdin path already uses.
var stdoutIsTTY = func() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}()

// colorEnabled gates ANSI escapes. Honors the NO_COLOR convention and dumb terminals.
var colorEnabled = stdoutIsTTY && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"

const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
	ansiItalic  = "\x1b[3m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiBlue    = "\x1b[34m"
	ansiMagenta = "\x1b[35m"
	ansiCyan    = "\x1b[36m"
)

func sgr(code, s string) string {
	if !colorEnabled {
		return s
	}
	return code + s + ansiReset
}

func bold(s string) string  { return sgr(ansiBold, s) }
func dim(s string) string   { return sgr(ansiDim, s) }
func green(s string) string { return sgr(ansiGreen, s) }

// termWidth returns the usable column count, from $COLUMNS when available.
func termWidth() int {
	if c := strings.TrimSpace(os.Getenv("COLUMNS")); c != "" {
		if n, err := strconv.Atoi(c); err == nil && n >= 40 {
			return n
		}
	}
	return 100
}

// ---- tables ----

type borderSet struct {
	tl, tm, tr string // top    ┌ ┬ ┐
	ml, mm, mr string // middle ├ ┼ ┤
	bl, bm, br string // bottom └ ┴ ┘
	h, v       string // ─ │
}

var boxBorders = borderSet{"┌", "┬", "┐", "├", "┼", "┤", "└", "┴", "┘", "─", "│"}
var asciiBorders = borderSet{"+", "+", "+", "+", "+", "+", "+", "+", "+", "-", "|"}

// renderTable lays out headers + rows as an aligned table. When the natural
// width overflows the terminal, the widest column is word-wrapped to fit. Box-
// drawing characters are used on a TTY, plain ASCII otherwise.
func renderTable(headers []string, rows [][]string) string {
	ncol := len(headers)
	if ncol == 0 {
		return ""
	}
	b := asciiBorders
	if colorEnabled {
		b = boxBorders
	}

	// Natural (unwrapped) width per column.
	width := make([]int, ncol)
	for i, h := range headers {
		width[i] = visLen(h)
	}
	for _, r := range rows {
		for i := 0; i < ncol && i < len(r); i++ {
			if w := maxLineLen(r[i]); w > width[i] {
				width[i] = w
			}
		}
	}

	// Shrink the widest column if the table overflows the terminal.
	// overhead = borders + " x " padding around each cell.
	overhead := ncol + 1 + 2*ncol
	avail := termWidth()
	for sum(width)+overhead > avail {
		wi := argmax(width)
		if width[wi] <= 24 {
			break
		}
		width[wi]--
	}

	var sb strings.Builder
	line := func(l, m, r string) {
		sb.WriteString(l)
		for i := 0; i < ncol; i++ {
			sb.WriteString(strings.Repeat(b.h, width[i]+2))
			if i < ncol-1 {
				sb.WriteString(m)
			}
		}
		sb.WriteString(r + "\n")
	}
	// A row may span multiple visual lines once cells are wrapped.
	emit := func(cells []string, styler func(string) string) {
		wrapped := make([][]string, ncol)
		height := 1
		for i := 0; i < ncol; i++ {
			var c string
			if i < len(cells) {
				c = cells[i]
			}
			wrapped[i] = wrapCell(c, width[i])
			if len(wrapped[i]) > height {
				height = len(wrapped[i])
			}
		}
		for row := 0; row < height; row++ {
			sb.WriteString(b.v)
			for i := 0; i < ncol; i++ {
				cell := ""
				if row < len(wrapped[i]) {
					cell = wrapped[i][row]
				}
				pad := width[i] - visLen(cell)
				if pad < 0 {
					pad = 0
				}
				content := cell
				if styler != nil {
					content = styler(cell)
				}
				sb.WriteString(" " + content + strings.Repeat(" ", pad) + " " + b.v)
			}
			sb.WriteString("\n")
		}
	}

	line(b.tl, b.tm, b.tr)
	emit(headers, bold)
	line(b.ml, b.mm, b.mr)
	for _, r := range rows {
		emit(r, nil)
	}
	line(b.bl, b.bm, b.br)
	return strings.TrimRight(sb.String(), "\n")
}

// wrapCell splits a possibly multi-line cell into display lines no wider than w.
func wrapCell(s string, w int) []string {
	var out []string
	for _, para := range strings.Split(s, "\n") {
		out = append(out, wrapText(para, w)...)
	}
	if len(out) == 0 {
		return []string{""}
	}
	return out
}

// wrapText word-wraps a single line to width w (soft-breaking overlong words).
func wrapText(s string, w int) []string {
	if w < 1 {
		w = 1
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	cur := ""
	for _, word := range words {
		for visLen(word) > w { // hard-break a word longer than the column
			if cur != "" {
				lines = append(lines, cur)
				cur = ""
			}
			lines = append(lines, sliceRunes(word, w))
			word = dropRunes(word, w)
		}
		switch {
		case cur == "":
			cur = word
		case visLen(cur)+1+visLen(word) <= w:
			cur += " " + word
		default:
			lines = append(lines, cur)
			cur = word
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

// ---- markdown ----

// renderMarkdown converts the optimizer's Markdown into styled terminal output.
// Headings become colored rules, pipe tables become aligned tables, fenced code
// blocks are boxed but their contents preserved verbatim (so a polished prompt
// stays copy-paste clean), and inline **bold** / `code` are highlighted.
func renderMarkdown(md string) string {
	if !colorEnabled {
		return md // plain terminals / pipes get the source untouched
	}
	lines := strings.Split(md, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Fenced code block — pass contents through unchanged inside a dim frame.
		if strings.HasPrefix(trimmed, "```") {
			lang := strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			var body []string
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				body = append(body, lines[i])
				i++
			}
			out = append(out, renderCodeBlock(lang, body))
			continue
		}

		// Pipe table — a header row followed by a |---|---| separator.
		if isTableRow(line) && i+1 < len(lines) && isTableSep(lines[i+1]) {
			headers := splitTableRow(line)
			var rows [][]string
			i += 2
			for i < len(lines) && isTableRow(lines[i]) {
				rows = append(rows, splitTableRow(lines[i]))
				i++
			}
			i--
			out = append(out, renderTable(headers, rows))
			continue
		}

		out = append(out, renderMarkdownLine(line))
	}
	return strings.Join(out, "\n")
}

func renderMarkdownLine(line string) string {
	trimmed := strings.TrimSpace(line)

	// Headings: colored bold + underline rule for h1/h2.
	if strings.HasPrefix(trimmed, "#") {
		level := 0
		for level < len(trimmed) && trimmed[level] == '#' {
			level++
		}
		text := strings.TrimSpace(trimmed[level:])
		text = stripInline(text)
		switch level {
		case 1:
			rule := strings.Repeat("═", minInt(visLen(text), termWidth()))
			return sgr(ansiBold+ansiCyan, text) + "\n" + dim(rule)
		case 2:
			rule := strings.Repeat("─", minInt(visLen(text), termWidth()))
			return sgr(ansiBold+ansiCyan, text) + "\n" + dim(rule)
		default:
			return sgr(ansiBold+ansiBlue, text)
		}
	}

	// Bullet lists: normalize marker to a colored bullet, keep indentation.
	if m := leadingBullet(line); m >= 0 {
		indent := line[:m]
		rest := strings.TrimSpace(line[m+1:])
		return indent + green("•") + " " + inlineFmt(rest)
	}

	// Horizontal rule.
	if trimmed == "---" || trimmed == "***" || trimmed == "___" {
		return dim(strings.Repeat("─", minInt(60, termWidth())))
	}

	return inlineFmt(line)
}

// renderCodeBlock frames a fenced block with a labeled top rule and a bottom
// rule, but leaves each content line FLUSH-LEFT and unstyled. No per-line border
// is added, so selecting the block copies the prompt verbatim — critical for the
// polished-prompt output, which users paste elsewhere.
func renderCodeBlock(lang string, body []string) string {
	w := termWidth()
	label := ""
	if lang != "" {
		label = " " + lang + " "
	}
	top := dim("╶─" + label + strings.Repeat("─", maxInt(4, w-visLen(label)-3)) + "╴")
	bottom := dim("╶" + strings.Repeat("─", w-2) + "╴")
	var sb strings.Builder
	sb.WriteString(top + "\n")
	for _, l := range body {
		sb.WriteString(l + "\n") // verbatim, copy-paste clean
	}
	sb.WriteString(bottom)
	return sb.String()
}

// inlineFmt applies **bold** and `code` styling within a line.
func inlineFmt(s string) string {
	s = replacePairs(s, "**", func(inner string) string { return bold(inner) })
	s = replacePairs(s, "`", func(inner string) string { return sgr(ansiYellow, inner) })
	return s
}

// stripInline removes markdown emphasis markers without adding color (used in
// headings, which are already colored as a whole).
func stripInline(s string) string {
	s = replacePairs(s, "**", func(inner string) string { return inner })
	s = replacePairs(s, "`", func(inner string) string { return inner })
	return s
}

// replacePairs replaces balanced `delim`-wrapped spans using fn.
func replacePairs(s, delim string, fn func(string) string) string {
	var sb strings.Builder
	for {
		start := strings.Index(s, delim)
		if start < 0 {
			sb.WriteString(s)
			break
		}
		end := strings.Index(s[start+len(delim):], delim)
		if end < 0 {
			sb.WriteString(s)
			break
		}
		inner := s[start+len(delim) : start+len(delim)+end]
		sb.WriteString(s[:start])
		sb.WriteString(fn(inner))
		s = s[start+len(delim)+end+len(delim):]
	}
	return sb.String()
}

// ---- small helpers ----

func isTableRow(s string) bool {
	t := strings.TrimSpace(s)
	return strings.HasPrefix(t, "|") && strings.Count(t, "|") >= 2
}

func isTableSep(s string) bool {
	t := strings.TrimSpace(s)
	if !strings.HasPrefix(t, "|") {
		return false
	}
	for _, r := range t {
		if r != '|' && r != '-' && r != ':' && r != ' ' {
			return false
		}
	}
	return strings.Contains(t, "-")
}

func splitTableRow(s string) []string {
	t := strings.TrimSpace(s)
	t = strings.TrimPrefix(t, "|")
	t = strings.TrimSuffix(t, "|")
	parts := strings.Split(t, "|")
	for i := range parts {
		parts[i] = stripInline(strings.TrimSpace(parts[i]))
	}
	return parts
}

// leadingBullet returns the index of a "- " or "* " list marker, or -1.
func leadingBullet(s string) int {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	if i+1 < len(s) && (s[i] == '-' || s[i] == '*') && s[i+1] == ' ' {
		return i
	}
	return -1
}

// visLen is the display width of s, ignoring ANSI escape sequences.
func visLen(s string) int {
	n, inEsc := 0, false
	for _, r := range s {
		switch {
		case inEsc:
			if r == 'm' {
				inEsc = false
			}
		case r == '\x1b':
			inEsc = true
		default:
			n++
		}
	}
	return n
}

func maxLineLen(s string) int {
	m := 0
	for _, l := range strings.Split(s, "\n") {
		if w := visLen(l); w > m {
			m = w
		}
	}
	return m
}

func sliceRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return string(r[:n])
}

func dropRunes(s string, n int) string {
	r := []rune(s)
	if n >= len(r) {
		return ""
	}
	return string(r[n:])
}

func sum(a []int) int {
	t := 0
	for _, v := range a {
		t += v
	}
	return t
}

func argmax(a []int) int {
	idx := 0
	for i, v := range a {
		if v > a[idx] {
			idx = i
		}
	}
	return idx
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
