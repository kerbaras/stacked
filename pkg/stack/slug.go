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
	nonAlnum  = regexp.MustCompile(`[^a-z0-9]+`)
	trimDash  = regexp.MustCompile(`^-+|-+$`)
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

// BranchName builds a branch name in the format stack/<stackName>/<nn>-<slug>.
func BranchName(stackName string, index int, title string) string {
	slug := Slugify(title)
	return fmt.Sprintf("stack/%s/%02d-%s", stackName, index, slug)
}
