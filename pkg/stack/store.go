package stack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const stateFile = "stacked.json"

// Store manages loading and saving stack state from .git/stacked.json.
type Store struct {
	State  *State
	gitDir string
}

// LoadStore reads stacked.json from the given git directory.
// If the file doesn't exist, returns a store with empty state.
func LoadStore(gitDir string) (*Store, error) {
	s := &Store{
		gitDir: gitDir,
		State: &State{
			Version: 1,
			Stacks:  make(map[string]*Stack),
		},
	}

	data, err := os.ReadFile(s.path())
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read stacked state: %w", err)
	}

	if err := json.Unmarshal(data, s.State); err != nil {
		return nil, fmt.Errorf("parse stacked state: %w", err)
	}

	if s.State.Stacks == nil {
		s.State.Stacks = make(map[string]*Stack)
	}

	return s, nil
}

// Save writes the current state to .git/stacked.json.
func (s *Store) Save() error {
	data, err := json.MarshalIndent(s.State, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal stacked state: %w", err)
	}
	if err := os.WriteFile(s.path(), data, 0o644); err != nil {
		return fmt.Errorf("write stacked state: %w", err)
	}
	return nil
}

// CurrentStack returns the active stack based on current_stack, or nil.
func (s *Store) CurrentStack() *Stack {
	return s.State.Stacks[s.State.CurrentStack]
}

// CurrentStackName returns the name of the current stack.
func (s *Store) CurrentStackName() string {
	return s.State.CurrentStack
}

// StackForBranch finds which stack contains the given branch.
// Returns the stack and its name, or nil/"" if not found.
func (s *Store) StackForBranch(branchName string) (*Stack, string) {
	for name, st := range s.State.Stacks {
		if st.HasBranch(branchName) {
			return st, name
		}
	}
	return nil, ""
}

// ResolveCurrentStack determines the current stack, first from current_stack,
// then by checking if branchName belongs to any known stack.
// Updates current_stack if found by branch lookup.
func (s *Store) ResolveCurrentStack(branchName string) (*Stack, string) {
	// Try current_stack first
	if st := s.CurrentStack(); st != nil {
		if st.HasBranch(branchName) {
			return st, s.State.CurrentStack
		}
	}

	// Fall back to scanning all stacks
	st, name := s.StackForBranch(branchName)
	if st != nil {
		s.State.CurrentStack = name
	}
	return st, name
}

func (s *Store) path() string {
	return filepath.Join(s.gitDir, stateFile)
}
