// Package stack provides the core data model for stacked pull requests.
package stack

import "time"

type Branch struct {
	Name      string    `json:"name"`
	Title     string    `json:"title"`
	PR        *int      `json:"pr,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Stack struct {
	Base     string   `json:"base"`
	Branches []Branch `json:"branches"`
}

type State struct {
	Version      int               `json:"version"`
	Stacks       map[string]*Stack `json:"stacks"`
	CurrentStack string            `json:"current_stack"`
}

// Parent returns the parent branch name (previous branch, or base for the first).
func (s *Stack) Parent(branchName string) string {
	for i, b := range s.Branches {
		if b.Name == branchName {
			if i == 0 {
				return s.Base
			}
			return s.Branches[i-1].Name
		}
	}
	return ""
}

// Descendants returns all branches above a given branch in the stack.
func (s *Stack) Descendants(branchName string) []Branch {
	for i, b := range s.Branches {
		if b.Name == branchName && i+1 < len(s.Branches) {
			return s.Branches[i+1:]
		}
	}
	return nil
}

// Tip returns the topmost branch in the stack, or nil if empty.
func (s *Stack) Tip() *Branch {
	if len(s.Branches) == 0 {
		return nil
	}
	return &s.Branches[len(s.Branches)-1]
}

// IndexOf returns the 0-based index of a branch, or -1 if not found.
func (s *Stack) IndexOf(branchName string) int {
	for i, b := range s.Branches {
		if b.Name == branchName {
			return i
		}
	}
	return -1
}

// HasBranch reports whether the stack contains a branch with the given name.
func (s *Stack) HasBranch(branchName string) bool {
	return s.IndexOf(branchName) >= 0
}
