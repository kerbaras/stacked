package ui

import (
	"fmt"
	"strings"
)

// BranchInfo holds display metadata for a single branch in the graph.
type BranchInfo struct {
	Name        string
	Title       string
	Hash        string // abbreviated commit hash
	PRNumber    *int
	IsCurrent   bool
	NeedsRebase bool
}

// RenderGraph renders a git-log-style graph for a stack.
// base is the stack's base branch name.
// verbose controls whether commit hashes are shown.
func RenderGraph(branches []BranchInfo, base string, verbose bool) string {
	var b strings.Builder

	// Render top-to-bottom (tip first)
	for i := len(branches) - 1; i >= 0; i-- {
		br := branches[i]

		// Node symbol
		if br.IsCurrent {
			b.WriteString(Yellow.Render("*"))
		} else {
			b.WriteString(Cyan.Render("*"))
		}
		b.WriteString(" ")

		// Optional hash
		if verbose && br.Hash != "" {
			b.WriteString(Dim.Render(br.Hash))
			b.WriteString(" ")
		}

		// Title
		b.WriteString(br.Title)

		// Decorations
		b.WriteString("  ")
		b.WriteString(renderDecorations(br))

		b.WriteString("\n")

		// Connector
		if i > 0 || base != "" {
			b.WriteString(Dim.Render("|"))
			b.WriteString("\n")
		}
	}

	// Base
	b.WriteString(Dim.Render("○"))
	b.WriteString(" ")
	b.WriteString(Dim.Render(base + " (base)"))
	b.WriteString("\n")

	return b.String()
}

func renderDecorations(br BranchInfo) string {
	var parts []string

	if br.IsCurrent {
		parts = append(parts, Green.Render("HEAD -> "+br.Name))
	} else {
		parts = append(parts, Cyan.Render(br.Name))
	}

	if br.PRNumber != nil {
		pr := fmt.Sprintf("PR #%d", *br.PRNumber)
		if br.NeedsRebase {
			pr += " " + Yellow.Render("⚠ needs rebase")
		}
		parts = append(parts, Dim.Render(pr))
	} else {
		parts = append(parts, Dim.Render("no PR"))
	}

	return "(" + strings.Join(parts, ", ") + ")"
}
