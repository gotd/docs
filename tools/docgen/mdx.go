package main

import "strings"

// mdxText makes an arbitrary string safe to embed in MDX prose (Docusaurus 3 /
// MDX v3, with mdx1 compatibility disabled). Angle brackets and braces would
// otherwise be parsed as JSX/expressions.
func mdxText(s string) string {
	r := strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"{", "&#123;",
		"}", "&#125;",
	)
	return r.Replace(s)
}

// mdxCell makes a string safe to embed in a Markdown table cell: MDX-safe plus
// escaped pipes and no line breaks.
func mdxCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return mdxText(s)
}
