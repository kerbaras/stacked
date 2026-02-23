package github

import (
	"fmt"
	"strings"
)

const (
	sentinelStart = "<!-- stacked:stack-diagram -->"
	sentinelEnd   = "<!-- /stacked:stack-diagram -->"
)

// DiagramBranch holds info needed to render one line of the stack diagram.
type DiagramBranch struct {
	Title    string
	PRNumber *int
	PRURL    string
	IsCurrent bool
}

// RenderDiagram produces a markdown stack diagram, rendered top-down.
// base is the stack's base branch name (e.g. "main").
func RenderDiagram(branches []DiagramBranch, base string) string {
	var b strings.Builder
	b.WriteString(sentinelStart + "\n")
	b.WriteString("> **Stack**:\n")

	for i := len(branches) - 1; i >= 0; i-- {
		br := branches[i]
		num := len(branches) - (len(branches) - 1 - i)

		b.WriteString("> ")
		b.WriteString(fmt.Sprintf("%d. ", num))

		if br.IsCurrent {
			b.WriteString(fmt.Sprintf("**%s**", br.Title))
			b.WriteString(" \u2190 *this PR*")
		} else if br.PRNumber != nil && br.PRURL != "" {
			b.WriteString(fmt.Sprintf("[%s](%s)", br.Title, br.PRURL))
		} else if br.PRNumber != nil {
			b.WriteString(fmt.Sprintf("%s (PR #%d)", br.Title, *br.PRNumber))
		} else {
			b.WriteString(fmt.Sprintf("%s *(not yet opened)*", br.Title))
		}

		b.WriteString("\n")
	}

	b.WriteString(">\n")
	b.WriteString(fmt.Sprintf("> \u2b07 Base: `%s`\n", base))
	b.WriteString(sentinelEnd + "\n")

	return b.String()
}

// UpdateBody replaces the diagram in a PR body, or appends it if not found.
func UpdateBody(body, diagram string) string {
	startIdx := strings.Index(body, sentinelStart)
	endIdx := strings.Index(body, sentinelEnd)

	if startIdx >= 0 && endIdx >= 0 {
		endIdx += len(sentinelEnd)
		// Include trailing newline if present
		if endIdx < len(body) && body[endIdx] == '\n' {
			endIdx++
		}
		return body[:startIdx] + diagram + body[endIdx:]
	}

	// No existing diagram — append
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	if body != "" {
		body += "\n"
	}
	return body + diagram
}
