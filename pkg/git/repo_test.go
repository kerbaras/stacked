package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTestRepo creates a git repo in a temp dir with one commit.
func initTestRepo(t *testing.T) *Repo {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}

	run("init", "-b", "main")
	run("commit", "--allow-empty", "-m", "initial commit")

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return repo
}

func TestCurrentBranch(t *testing.T) {
	repo := initTestRepo(t)

	branch, err := repo.CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if branch != "main" {
		t.Errorf("CurrentBranch = %q, want %q", branch, "main")
	}
}

func TestBranchExists(t *testing.T) {
	repo := initTestRepo(t)

	if !repo.BranchExists("main") {
		t.Error("BranchExists(main) should be true")
	}
	if repo.BranchExists("nonexistent") {
		t.Error("BranchExists(nonexistent) should be false")
	}
}

func TestCheckout(t *testing.T) {
	repo := initTestRepo(t)

	if err := repo.Checkout("feature", true); err != nil {
		t.Fatalf("Checkout create: %v", err)
	}

	branch, err := repo.CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if branch != "feature" {
		t.Errorf("after checkout -b, branch = %q, want %q", branch, "feature")
	}

	if err := repo.Checkout("main", false); err != nil {
		t.Fatalf("Checkout existing: %v", err)
	}

	branch, _ = repo.CurrentBranch()
	if branch != "main" {
		t.Errorf("after checkout, branch = %q, want %q", branch, "main")
	}
}

func TestBranchTip(t *testing.T) {
	repo := initTestRepo(t)

	hash, err := repo.BranchTip("main")
	if err != nil {
		t.Fatalf("BranchTip: %v", err)
	}
	if len(hash) != 7 {
		t.Errorf("BranchTip hash length = %d, want 7", len(hash))
	}
}

func TestRevParse(t *testing.T) {
	repo := initTestRepo(t)

	hash, err := repo.RevParse("main")
	if err != nil {
		t.Fatalf("RevParse: %v", err)
	}
	if len(hash) != 40 {
		t.Errorf("RevParse hash length = %d, want 40", len(hash))
	}
}

func TestGitDir(t *testing.T) {
	repo := initTestRepo(t)

	gitDir := repo.GitDir()
	if !strings.HasSuffix(gitDir, ".git") {
		t.Errorf("GitDir = %q, should end with .git", gitDir)
	}
}

func TestPath(t *testing.T) {
	repo := initTestRepo(t)

	if repo.Path() == "" {
		t.Error("Path should not be empty")
	}
}

func TestNeedsRebase(t *testing.T) {
	repo := initTestRepo(t)

	// Create a feature branch
	if err := repo.Checkout("feature", true); err != nil {
		t.Fatal(err)
	}

	// Add a commit to feature
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", "feature commit")
	cmd.Dir = repo.Path()
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("commit: %v\n%s", err, out)
	}

	// main should NOT need rebase relative to feature (main is behind)
	needs, err := repo.NeedsRebase("feature", "main")
	if err != nil {
		t.Fatalf("NeedsRebase: %v", err)
	}
	if needs {
		t.Error("feature should not need rebase (main has no new commits)")
	}

	// Go back to main, add a commit
	if err := repo.Checkout("main", false); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "main commit")
	cmd.Dir = repo.Path()
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("commit: %v\n%s", err, out)
	}

	// Now feature needs rebase (main has new commits)
	needs, err = repo.NeedsRebase("feature", "main")
	if err != nil {
		t.Fatalf("NeedsRebase: %v", err)
	}
	if !needs {
		t.Error("feature should need rebase (main has new commits)")
	}
}

func TestInstallHooks(t *testing.T) {
	repo := initTestRepo(t)

	if repo.HasHooks() {
		t.Error("should not have hooks initially")
	}

	if err := repo.InstallHooks(); err != nil {
		t.Fatalf("InstallHooks: %v", err)
	}

	if !repo.HasHooks() {
		t.Error("should have hooks after install")
	}

	// Check hook files exist and are executable
	for _, name := range []string{"post-merge", "post-checkout"} {
		hookPath := filepath.Join(repo.GitDir(), "hooks", name)
		info, err := os.Stat(hookPath)
		if err != nil {
			t.Errorf("hook %s not found: %v", name, err)
			continue
		}
		if info.Mode()&0o111 == 0 {
			t.Errorf("hook %s is not executable", name)
		}
	}

	// Install again should be idempotent
	if err := repo.InstallHooks(); err != nil {
		t.Fatalf("second InstallHooks: %v", err)
	}
}

func TestDeleteBranch(t *testing.T) {
	repo := initTestRepo(t)

	if err := repo.Checkout("to-delete", true); err != nil {
		t.Fatal(err)
	}
	if err := repo.Checkout("main", false); err != nil {
		t.Fatal(err)
	}

	if err := repo.DeleteBranch("to-delete"); err != nil {
		t.Fatalf("DeleteBranch: %v", err)
	}

	if repo.BranchExists("to-delete") {
		t.Error("branch should be deleted")
	}
}
