package github

import (
	"strings"
	"testing"
)

func intPtr(n int) *int { return &n }

func TestRenderDiagram(t *testing.T) {
	branches := []DiagramBranch{
		{Title: "add auth", PRNumber: intPtr(123), PRURL: "https://github.com/o/r/pull/123"},
		{Title: "add tokens", PRNumber: intPtr(124), PRURL: "https://github.com/o/r/pull/124", IsCurrent: true},
		{Title: "add logout", PRNumber: nil},
	}

	got := RenderDiagram(branches, "main")

	// Check sentinels
	if !strings.Contains(got, sentinelStart) {
		t.Error("missing start sentinel")
	}
	if !strings.Contains(got, sentinelEnd) {
		t.Error("missing end sentinel")
	}

	// Check order (top-down: tip first)
	logoutIdx := strings.Index(got, "add logout")
	tokensIdx := strings.Index(got, "add tokens")
	authIdx := strings.Index(got, "add auth")
	if logoutIdx > tokensIdx || tokensIdx > authIdx {
		t.Error("branches should be rendered top-down (tip first)")
	}

	// Current PR bolded
	if !strings.Contains(got, "**add tokens**") {
		t.Error("current PR should be bolded")
	}
	if !strings.Contains(got, "← *this PR*") {
		t.Error("current PR should have marker")
	}

	// PR with URL is a link
	if !strings.Contains(got, "[add auth](https://github.com/o/r/pull/123)") {
		t.Error("PR with URL should be a link")
	}

	// No PR shows "not yet opened"
	if !strings.Contains(got, "*(not yet opened)*") {
		t.Error("branch without PR should show 'not yet opened'")
	}

	// Base
	if !strings.Contains(got, "`main`") {
		t.Error("should show base branch")
	}
}

func TestUpdateBody_Insert(t *testing.T) {
	body := "Some PR description\n\nMore details here."
	diagram := RenderDiagram([]DiagramBranch{{Title: "test", IsCurrent: true}}, "main")

	result := UpdateBody(body, diagram)

	if !strings.Contains(result, "Some PR description") {
		t.Error("should preserve original body")
	}
	if !strings.Contains(result, sentinelStart) {
		t.Error("should contain diagram")
	}
}

func TestUpdateBody_Replace(t *testing.T) {
	oldDiagram := RenderDiagram([]DiagramBranch{{Title: "old", IsCurrent: true}}, "main")
	body := "Description\n\n" + oldDiagram + "\nFooter"

	newDiagram := RenderDiagram([]DiagramBranch{{Title: "new", IsCurrent: true}}, "main")
	result := UpdateBody(body, newDiagram)

	if strings.Contains(result, "old") {
		t.Error("should replace old diagram")
	}
	if !strings.Contains(result, "new") {
		t.Error("should contain new diagram")
	}
	if !strings.Contains(result, "Description") {
		t.Error("should preserve text before diagram")
	}
	if !strings.Contains(result, "Footer") {
		t.Error("should preserve text after diagram")
	}
}

func TestUpdateBody_Idempotent(t *testing.T) {
	diagram := RenderDiagram([]DiagramBranch{
		{Title: "auth", PRNumber: intPtr(1), PRURL: "https://github.com/o/r/pull/1"},
		{Title: "tokens", IsCurrent: true},
	}, "main")

	body := "Description\n\n" + diagram
	result := UpdateBody(body, diagram)
	result2 := UpdateBody(result, diagram)

	if result != result2 {
		t.Error("UpdateBody should be idempotent")
	}
}
