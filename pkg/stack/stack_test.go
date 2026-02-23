package stack

import (
	"testing"
	"time"
)

func testStack() *Stack {
	return &Stack{
		Base: "main",
		Branches: []Branch{
			{Name: "stack/feat/01-auth", Title: "add auth", CreatedAt: time.Now()},
			{Name: "stack/feat/02-tokens", Title: "add tokens", CreatedAt: time.Now()},
			{Name: "stack/feat/03-logout", Title: "add logout", CreatedAt: time.Now()},
		},
	}
}

func TestParent(t *testing.T) {
	tests := []struct {
		name   string
		branch string
		want   string
	}{
		{"first branch returns base", "stack/feat/01-auth", "main"},
		{"second branch returns first", "stack/feat/02-tokens", "stack/feat/01-auth"},
		{"third branch returns second", "stack/feat/03-logout", "stack/feat/02-tokens"},
		{"unknown branch returns empty", "stack/feat/99-nope", ""},
	}

	s := testStack()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Parent(tt.branch)
			if got != tt.want {
				t.Errorf("Parent(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}

func TestDescendants(t *testing.T) {
	s := testStack()

	tests := []struct {
		name   string
		branch string
		want   int
	}{
		{"from first", "stack/feat/01-auth", 2},
		{"from second", "stack/feat/02-tokens", 1},
		{"from last", "stack/feat/03-logout", 0},
		{"unknown", "stack/feat/99-nope", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Descendants(tt.branch)
			if len(got) != tt.want {
				t.Errorf("Descendants(%q) returned %d branches, want %d", tt.branch, len(got), tt.want)
			}
		})
	}
}

func TestTip(t *testing.T) {
	s := testStack()
	tip := s.Tip()
	if tip == nil || tip.Name != "stack/feat/03-logout" {
		t.Errorf("Tip() = %v, want stack/feat/03-logout", tip)
	}

	empty := &Stack{Base: "main"}
	if empty.Tip() != nil {
		t.Error("Tip() on empty stack should return nil")
	}
}

func TestIndexOf(t *testing.T) {
	s := testStack()

	tests := []struct {
		branch string
		want   int
	}{
		{"stack/feat/01-auth", 0},
		{"stack/feat/02-tokens", 1},
		{"stack/feat/03-logout", 2},
		{"stack/feat/99-nope", -1},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			got := s.IndexOf(tt.branch)
			if got != tt.want {
				t.Errorf("IndexOf(%q) = %d, want %d", tt.branch, got, tt.want)
			}
		})
	}
}

func TestHasBranch(t *testing.T) {
	s := testStack()

	if !s.HasBranch("stack/feat/01-auth") {
		t.Error("HasBranch should return true for existing branch")
	}
	if s.HasBranch("stack/feat/99-nope") {
		t.Error("HasBranch should return false for unknown branch")
	}
}
