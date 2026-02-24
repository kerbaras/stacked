package ui

import (
	"fmt"
	"os"
)

// Step prints a styled step message to stderr (e.g. "● fetching origin").
func Step(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Cyan.Render("●"), msg)
}

// Stepf prints a formatted styled step message to stderr.
func Stepf(format string, a ...any) {
	Step(fmt.Sprintf(format, a...))
}

// Success prints a styled success message to stderr (e.g. "✓ all branches pushed").
func Success(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Green.Render("✓"), msg)
}

// Successf prints a formatted styled success message to stderr.
func Successf(format string, a ...any) {
	Success(fmt.Sprintf(format, a...))
}

// Warn prints a styled warning to stderr (e.g. "⚠ could not push branch").
func Warn(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Yellow.Render("⚠"), msg)
}

// Warnf prints a formatted styled warning to stderr.
func Warnf(format string, a ...any) {
	Warn(fmt.Sprintf(format, a...))
}

// Error prints a styled error message to stderr.
func Error(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Red.Render("✗"), msg)
}

// Errorf prints a formatted styled error message to stderr.
func Errorf(format string, a ...any) {
	Error(fmt.Sprintf(format, a...))
}

// Hint prints a styled hint to stderr (e.g. "hint: run `stacked sync`").
func Hint(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Dim.Render("hint:"), msg)
}

// Hintf prints a formatted styled hint to stderr.
func Hintf(format string, a ...any) {
	Hint(fmt.Sprintf(format, a...))
}

// Info prints an informational message to stderr.
func Info(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Dim.Render("─"), msg)
}

// Infof prints a formatted informational message to stderr.
func Infof(format string, a ...any) {
	Info(fmt.Sprintf(format, a...))
}

// BranchName renders a branch name with the branch style.
func BranchName(name string) string {
	return Cyan.Render(name)
}

// PRRef renders a PR reference like "PR #123".
func PRRef(number int) string {
	return Green.Render(fmt.Sprintf("PR #%d", number))
}

// Faint renders dimmed text.
func Faint(s string) string {
	return Dim.Render(s)
}
