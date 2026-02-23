package cmd

import (
	"fmt"
	"os"

	"github.com/kerbaras/stacked/pkg/git"
	"github.com/kerbaras/stacked/pkg/stack"
)

func openRepo() (*git.Repo, error) {
	repo, err := git.Open(".")
	if err != nil {
		return nil, fmt.Errorf("not a git repository (or any parent): %w", err)
	}
	return repo, nil
}

func loadStore(repo *git.Repo) (*stack.Store, error) {
	return stack.LoadStore(repo.GitDir())
}

// mustCurrentStack resolves the current stack based on HEAD.
// Returns the stack, its name, and the current branch.
func mustCurrentStack(repo *git.Repo, store *stack.Store) (*stack.Stack, string, string, error) {
	branch, err := repo.CurrentBranch()
	if err != nil {
		return nil, "", "", err
	}

	st, name := store.ResolveCurrentStack(branch)
	if st == nil {
		return nil, "", "", fmt.Errorf("no stack found for branch %q; run `stacked new` to create one", branch)
	}

	return st, name, branch, nil
}

func syncHint(repo *git.Repo, st *stack.Stack) {
	local, err1 := repo.RevParse(st.Base)
	remote, err2 := repo.RevParse("origin/" + st.Base)
	if err1 != nil || err2 != nil {
		return
	}
	if local != remote {
		fmt.Fprintf(os.Stderr, "hint: %s has new commits. Run `stacked sync` to update.\n", st.Base)
	}
}
