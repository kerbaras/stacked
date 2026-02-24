package stack

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const maxSlugLen = 50

var (
	nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
	trimDash = regexp.MustCompile(`^-+|-+$`)
)

// Slugify converts a title into a URL/branch-safe slug.
// It strips non-ASCII, lowercases, replaces non-alphanumeric runs with dashes,
// and truncates to maxSlugLen characters.
func Slugify(title string) string {
	// NFC normalize, then strip non-ASCII
	s := norm.NFC.String(title)
	var b strings.Builder
	for _, r := range s {
		if r < unicode.MaxASCII {
			b.WriteRune(r)
		}
	}
	s = strings.ToLower(b.String())

	// Replace non-alphanumeric runs with single dash
	s = nonAlnum.ReplaceAllString(s, "-")
	s = trimDash.ReplaceAllString(s, "")

	if len(s) > maxSlugLen {
		s = s[:maxSlugLen]
		s = trimDash.ReplaceAllString(s, "")
	}

	return s
}

// StackName derives a stack name from the first branch title.
// It uses the full slug of the title.
func StackName(title string) string {
	return Slugify(title)
}

var indexPattern = regexp.MustCompile(`%0?\d*d`)

// BranchName builds a branch name from a template.
// Supported placeholders: %name (stack name), %slug (slugified title),
// and printf-style integer verbs like %02d (branch index).
func BranchName(template, stackName string, index int, title string) string {
	slug := Slugify(title)
	s := strings.ReplaceAll(template, "%name", stackName)
	s = strings.ReplaceAll(s, "%slug", slug)
	s = indexPattern.ReplaceAllStringFunc(s, func(match string) string {
		return fmt.Sprintf(match, index)
	})
	return s
}
