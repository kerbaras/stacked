package stack

import "time"

type Branch struct {
	Name      string    `json:"name"`
	PR        *int      `json:"pr,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Stack struct {
	Base     string   `json:"base"`
	Branches []Branch `json:"branches"`
}

type State struct {
	Stacks       map[string]*Stack `json:"stacks"`
	CurrentStack string            `json:"current_stack"`
}

// Returns the parent branch name for a given branch in the stack
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

// Returns all branches that need rebasing if a given branch changes
func (s *Stack) Descendants(branchName string) []Branch {
	for i, b := range s.Branches {
		if b.Name == branchName && i+1 < len(s.Branches) {
			return s.Branches[i+1:]
		}
	}
	return nil
}
