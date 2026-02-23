package git

import (
	"os"
	"os/exec"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Repo struct {
	repo *gogit.Repository
	path string
}

func Open(path string) (*Repo, error) {
	r, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, err
	}
	return &Repo{repo: r, path: path}, nil
}

// Reads via go-git (fast, no subprocess)
func (r *Repo) CurrentBranch() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", err
	}
	return head.Name().Short(), nil
}

func (r *Repo) BranchExists(name string) bool {
	_, err := r.repo.Reference(
		plumbing.NewBranchReferenceName(name), true,
	)
	return err == nil
}

// Mutations via exec (reliable)
func (r *Repo) Checkout(branch string, create bool) error {
	args := []string{"checkout"}
	if create {
		args = append(args, "-b")
	}
	args = append(args, branch)
	return r.exec(args...)
}

func (r *Repo) RebaseOnto(newBase, oldBase, branch string) error {
	return r.exec("rebase", "--onto", newBase, oldBase, branch)
}

func (r *Repo) Push(branch string, force bool) error {
	args := []string{"push", "origin", branch}
	if force {
		args = []string{"push", "--force-with-lease", "origin", branch}
	}
	return r.exec(args...)
}

func (r *Repo) exec(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
