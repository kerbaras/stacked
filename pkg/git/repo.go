// Package git provides git operations using go-git for reads and exec for mutations.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Repo wraps a git repository for both go-git reads and exec-based mutations.
type Repo struct {
	repo *gogit.Repository
	path string
}

// Open opens a git repository at the given path, detecting .git upward.
func Open(path string) (*Repo, error) {
	r, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open git repo: %w", err)
	}

	wt, err := r.Worktree()
	if err != nil {
		return nil, fmt.Errorf("get worktree: %w", err)
	}

	return &Repo{repo: r, path: wt.Filesystem.Root()}, nil
}

// Path returns the working directory root.
func (r *Repo) Path() string {
	return r.path
}

// GitDir returns the path to the .git directory.
func (r *Repo) GitDir() string {
	return filepath.Join(r.path, ".git")
}

// CurrentBranch returns the short name of HEAD, or an error if detached.
func (r *Repo) CurrentBranch() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("resolve HEAD: %w", err)
	}
	if !head.Name().IsBranch() {
		return "", fmt.Errorf("HEAD is detached at %s; checkout a branch first", head.Hash().String()[:8])
	}
	return head.Name().Short(), nil
}

// BranchExists reports whether a local branch with the given name exists.
func (r *Repo) BranchExists(name string) bool {
	_, err := r.repo.Reference(plumbing.NewBranchReferenceName(name), true)
	return err == nil
}

// BranchTip returns the abbreviated commit hash for a branch.
func (r *Repo) BranchTip(name string) (string, error) {
	ref, err := r.repo.Reference(plumbing.NewBranchReferenceName(name), true)
	if err != nil {
		return "", fmt.Errorf("resolve branch %s: %w", name, err)
	}
	return ref.Hash().String()[:7], nil
}

// RevParse resolves any ref to a full commit hash string.
func (r *Repo) RevParse(ref string) (string, error) {
	out, err := r.execCapture("rev-parse", ref)
	if err != nil {
		return "", fmt.Errorf("rev-parse %s: %w", ref, err)
	}
	return strings.TrimSpace(out), nil
}

// Checkout switches to a branch. If create is true, creates it first.
func (r *Repo) Checkout(branch string, create bool) error {
	args := []string{"checkout"}
	if create {
		args = append(args, "-b")
	}
	args = append(args, branch)
	return r.exec(args...)
}

// Push pushes a branch to origin. If force is true, uses --force-with-lease.
func (r *Repo) Push(branch string, force bool) error {
	args := []string{"push", "origin", branch}
	if force {
		args = []string{"push", "--force-with-lease", "origin", branch}
	}
	return r.exec(args...)
}

// CheckoutSilent switches to a branch without printing output.
func (r *Repo) CheckoutSilent(branch string, create bool) error {
	args := []string{"checkout"}
	if create {
		args = append(args, "-b")
	}
	args = append(args, branch)
	return r.execSilent(args...)
}

// PushSilent pushes a branch to origin without printing output.
func (r *Repo) PushSilent(branch string, force bool) error {
	args := []string{"push", "origin", branch}
	if force {
		args = []string{"push", "--force-with-lease", "origin", branch}
	}
	return r.execSilent(args...)
}

func (r *Repo) exec(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *Repo) execSilent(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	if out, err := cmd.CombinedOutput(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%s: %s", ee, strings.TrimSpace(string(out)))
		}
		return err
	}
	return nil
}

func (r *Repo) execCapture(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s: %s", err, string(ee.Stderr))
		}
		return "", err
	}
	return string(out), nil
}
