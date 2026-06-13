package main

import (
	"regexp"
	"strings"
)

// docInfo is the structured form of a method's doc comment.
type docInfo struct {
	tlName  string
	hash    string
	summary string
	url     string
	errors  []ErrorRow
}

var (
	// "MessagesSendMessage invokes method messages.sendMessage#fef48f62 returning error if any."
	reInvokes = regexp.MustCompile(`invokes method ([a-zA-Z][\w.]*)#([0-9a-fA-F]+)`)
	// "See https://core.telegram.org/method/messages.sendMessage for reference."
	reSeeURL = regexp.MustCompile(`See (https?://\S+) for reference`)
	// "400 CHANNEL_INVALID: The provided channel is invalid."
	reError = regexp.MustCompile(`^(\d+)\s+([A-Z0-9_%]+):\s*(.*)$`)
)

// parseDoc extracts structured information from a method doc comment.
// text is the comment with // markers already stripped (ast.CommentGroup.Text).
func parseDoc(text string) docInfo {
	var d docInfo
	if m := reInvokes.FindStringSubmatch(text); m != nil {
		d.tlName = m[1]
		d.hash = m[2]
	}
	if d.tlName == "" {
		return d
	}
	if m := reSeeURL.FindStringSubmatch(text); m != nil {
		d.url = m[1]
	}

	lines := strings.Split(text, "\n")
	var summary []string
	inErrors := false
	inLinks := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 {
			continue // the "invokes method" line
		}
		switch {
		case strings.HasPrefix(trimmed, "Possible errors:"):
			inErrors, inLinks = true, false
			continue
		case strings.HasPrefix(trimmed, "Links:"):
			inLinks = true
			continue
		case strings.HasPrefix(trimmed, "See ") && strings.Contains(trimmed, "for reference"):
			continue
		}
		if inErrors {
			if m := reError.FindStringSubmatch(trimmed); m != nil {
				d.errors = append(d.errors, ErrorRow{Code: m[1], Name: m[2], Description: strings.TrimSpace(m[3])})
			}
			continue
		}
		if inLinks {
			// Footnote link lines like " 1) https://..."; end the block on a blank line.
			if trimmed == "" {
				inLinks = false
			}
			continue
		}
		if trimmed != "" {
			summary = append(summary, trimmed)
		}
	}
	d.summary = cleanInline(strings.Join(summary, " "))
	return d
}

var (
	// "InputPeerUser represents TL type `inputPeerUser#dde8a54c`."
	reTLType = regexp.MustCompile("represents TL type `([\\w.]+)#([0-9a-fA-F]+)`")
	// "InputPeerClass represents InputPeer generic type."
	reGeneric = regexp.MustCompile(`represents ([\w.]+) generic type`)
)

// typeDoc is the structured form of a constructor or class doc comment.
type typeDoc struct {
	tlName  string
	hash    string
	summary string
	url     string
}

// parseTypeDoc parses a constructor struct doc comment.
func parseTypeDoc(text string) typeDoc {
	var d typeDoc
	m := reTLType.FindStringSubmatch(text)
	if m == nil {
		return d
	}
	d.tlName, d.hash = m[1], m[2]
	if u := reSeeURL.FindStringSubmatch(text); u != nil {
		d.url = u[1]
	}
	d.summary = docBody(text)
	return d
}

// parseClassDoc parses a TL class (interface) doc comment.
func parseClassDoc(text string) typeDoc {
	var d typeDoc
	m := reGeneric.FindStringSubmatch(text)
	if m == nil {
		return d
	}
	d.tlName = m[1]
	if u := reSeeURL.FindStringSubmatch(text); u != nil {
		d.url = u[1]
	}
	d.summary = docBody(text)
	return d
}

// docBody returns the first prose paragraph of a doc comment (the lines after
// the header line, up to the first blank line), ignoring section blocks like
// Links/See/Constructors/Example. Class docs have no prose paragraph and yield "".
func docBody(text string) string {
	lines := strings.Split(text, "\n")
	var body []string
	for i := 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			if len(body) > 0 {
				break // end of the first paragraph
			}
			continue
		}
		if isSectionHeader(trimmed) {
			break
		}
		body = append(body, trimmed)
	}
	return cleanInline(strings.Join(body, " "))
}

func isSectionHeader(line string) bool {
	switch {
	case strings.HasPrefix(line, "Links:"),
		strings.HasPrefix(line, "Constructors:"),
		strings.HasPrefix(line, "Example:"),
		strings.HasPrefix(line, "See ") && strings.Contains(line, "for reference"):
		return true
	}
	return false
}

var (
	reHelperNote = regexp.MustCompile(`(?m)^Use \w+ and \w+ helpers\.$`)
	reSuperscript = regexp.MustCompile(`[\x{00B9}\x{00B2}\x{00B3}\x{2070}-\x{2079}]+`)
	reWhitespace  = regexp.MustCompile(`\s+`)
)

// hasHelperNote reports whether a field doc marks an optional (conditional) field.
func hasHelperNote(doc string) bool {
	return reHelperNote.MatchString(doc)
}

// cleanInline removes footnote superscripts and collapses all whitespace to
// single spaces, yielding a one-line string.
func cleanInline(s string) string {
	s = reSuperscript.ReplaceAllString(s, "")
	s = reWhitespace.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// cleanDescription turns a field/struct doc comment into a single-line,
// table-safe description: it drops the trailing "Links:" footnote block and
// "Use ... helpers." note, removes footnote superscripts, and collapses
// whitespace.
func cleanDescription(doc string) string {
	if i := strings.Index(doc, "\nLinks:"); i >= 0 {
		doc = doc[:i]
	}
	doc = reHelperNote.ReplaceAllString(doc, "")
	return cleanInline(doc)
}

// yamlString quotes a string for safe use as a YAML front-matter scalar value.
func yamlString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

// kebab converts a camelCase TL short name to kebab-case (sendMessage -> send-message).
func kebab(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r - 'A' + 'a')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
