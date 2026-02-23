package git

import (
	"fmt"
	"strconv"
	"strings"
)

// Rebase rebases branch onto parent using `git rebase --empty=drop <parent> <branch>`.
func (r *Repo) Rebase(parent, branch string) error {
	return r.exec("rebase", "--empty=drop", parent, branch)
}

// RebaseContinue continues a paused rebase.
func (r *Repo) RebaseContinue() error {
	return r.exec("rebase", "--continue")
}

// RebaseAbort aborts a paused rebase.
func (r *Repo) RebaseAbort() error {
	return r.exec("rebase", "--abort")
}

// Fetch fetches from a remote.
func (r *Repo) Fetch(remote string) error {
	return r.exec("fetch", remote)
}

// Pull pulls the current branch with fast-forward only.
func (r *Repo) Pull(remote, branch string) error {
	return r.exec("pull", "--ff-only", remote, branch)
}

// DeleteBranch force-deletes a local branch.
func (r *Repo) DeleteBranch(name string) error {
	return r.exec("branch", "-D", name)
}

// NeedsRebase reports whether parent has commits not in branch.
func (r *Repo) NeedsRebase(branch, parent string) (bool, error) {
	out, err := r.execCapture("rev-list", "--count", fmt.Sprintf("%s..%s", branch, parent))
	if err != nil {
		return false, fmt.Errorf("check rebase needed %s..%s: %w", branch, parent, err)
	}
	count, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return false, fmt.Errorf("parse rev-list count: %w", err)
	}
	return count > 0, nil
}

// IsRebasing reports whether a rebase is currently in progress.
func (r *Repo) IsRebasing() bool {
	_, err := r.execCapture("rev-parse", "--verify", "REBASE_HEAD")
	return err == nil
}
