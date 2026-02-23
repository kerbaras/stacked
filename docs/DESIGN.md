# `stacked` — A CLI for Stacked Pull Requests

## Project Overview

`stacked` is a Go CLI tool that manages stacked pull requests — a workflow where a large feature is broken into a chain of small, dependent PRs that each target the previous branch in the stack rather than `main`. This makes code review faster, keeps diffs small, and lets developers keep shipping while waiting on reviews.

The tool manages the full lifecycle: creating stacks, navigating between branches, cascading rebases when a branch changes, pushing all branches, creating GitHub PRs with correct base branches, rendering stack diagrams in PR bodies, and syncing state after merges.

### Design Philosophy

- **Feel like git, not a wrapper.** Output should look like `git log --graph`. No proprietary visual language.
- **go-git for reads, `git` exec for mutations.** go-git has no rebase support. Shelling out to `git` for checkout, rebase, push is the pragmatic choice. Use go-git for fast reads like current branch, ref resolution, branch existence.
- **Offline-first, sync on demand.** No background daemons or server components. Every command does a cheap staleness check against `origin/<base>` and prompts the user to sync if needed.
- **Charm ecosystem for terminal UI.** Use `charmbracelet/lipgloss` for styling, `charmbracelet/tree` for graph rendering, and `charmbracelet/bubbletea` if interactive features are added later.

---

## CLI Interface

### Core Commands

```bash
# Start a new stack branching from a base
stacked new --from main "feat(svc): a"

# Add the next branch on top of the current stack tip
stacked next "feat(svc): b"
stacked next "feat(svc): c"

# Show the stack graph (git log --graph style)
stacked status
stacked status -v  # include abbreviated commit hashes

# Push all branches in the current stack (force-with-lease)
stacked push

# Create/update GitHub PRs for all branches in the stack
stacked review

# Sync after merges: prune merged branches, cascade rebase, retarget PRs
stacked sync

# Navigate the stack
stacked up        # checkout the branch above current in the stack
stacked down      # checkout the branch below current in the stack
stacked goto <n>  # checkout the nth branch in the stack (1-indexed)

# Rebase the entire stack (cascade from base upward)
stacked rebase

# Continue after resolving rebase conflicts
stacked continue

# Abort an in-progress rebase
stacked abort

# Install git hooks for auto-sync hints
stacked init --hooks

# List all stacks in the repo
stacked list
```

### Branch Naming

Auto-generate branch names from the description argument using a slug pattern:

```
stack/<stack-name>/<nn>-<slug>
```

Example: `stacked new --from main "feat(svc): add user auth"` creates:
- Stack name: `feat-svc-add-user-auth` (derived from first branch)
- Branch: `stack/feat-svc-add-user-auth/01-add-user-auth`

Subsequent `stacked next "feat(svc): add token refresh"` creates:
- Branch: `stack/feat-svc-add-user-auth/02-add-token-refresh`

The `<nn>` prefix keeps branches ordered in listings. The stack name is derived from the first branch's description by slugifying it.

---

## Data Model

### Stack Metadata — `.git/stacked.json`

Store stack state inside `.git/` to avoid polluting the working tree or commits. This file is the source of truth for stack structure.

```json
{
  "version": 1,
  "stacks": {
    "feat-svc-add-user-auth": {
      "base": "main",
      "branches": [
        {
          "name": "stack/feat-svc-add-user-auth/01-add-user-auth",
          "title": "feat(svc): add user auth",
          "pr": 123,
          "created_at": "2025-01-15T10:00:00Z"
        },
        {
          "name": "stack/feat-svc-add-user-auth/02-add-token-refresh",
          "title": "feat(svc): add token refresh",
          "pr": 124,
          "created_at": "2025-01-15T10:30:00Z"
        },
        {
          "name": "stack/feat-svc-add-user-auth/03-add-logout",
          "title": "feat(svc): add logout",
          "pr": null,
          "created_at": "2025-01-15T11:00:00Z"
        }
      ]
    }
  },
  "current_stack": "feat-svc-add-user-auth"
}
```

### Core Types

```go
// pkg/stack/stack.go
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

// Parent returns the parent branch name (the branch below in the stack, or base)
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

// Descendants returns all branches above a given branch in the stack
func (s *Stack) Descendants(branchName string) []Branch {
    for i, b := range s.Branches {
        if b.Name == branchName && i+1 < len(s.Branches) {
            return s.Branches[i+1:]
        }
    }
    return nil
}

// Tip returns the topmost branch in the stack
func (s *Stack) Tip() *Branch {
    if len(s.Branches) == 0 {
        return nil
    }
    return &s.Branches[len(s.Branches)-1]
}

// IndexOf returns the 0-based index of a branch, or -1
func (s *Stack) IndexOf(branchName string) int {
    for i, b := range s.Branches {
        if b.Name == branchName {
            return i
        }
    }
    return -1
}
```

---

## Project Structure

```
stacked/
├── cmd/
│   ├── root.go              # cobra root command, subcommand registration
│   ├── new.go               # `stacked new`
│   ├── next.go              # `stacked next`
│   ├── status.go            # `stacked status` (graph rendering)
│   ├── push.go              # `stacked push`
│   ├── review.go            # `stacked review` (create/update PRs)
│   ├── sync.go              # `stacked sync` (post-merge cleanup)
│   ├── rebase.go            # `stacked rebase` (cascade rebase)
│   ├── navigate.go          # `stacked up/down/goto`
│   ├── list.go              # `stacked list`
│   ├── init.go              # `stacked init` (hooks, setup)
│   └── continue_abort.go    # `stacked continue/abort`
├── pkg/
│   ├── stack/                   # core stack data model and operations
│   │   ├── stack.go             # Stack, Branch types, methods
│   │   ├── store.go             # load/save .git/stacked.json
│   │   └── slug.go              # branch name / stack name generation
│   ├── git/                     # git abstraction layer
│   │   ├── repo.go              # go-git for reads (current branch, ref resolution, etc.)
│   │   ├── exec.go              # shelling out to git for mutations
│   │   └── hooks.go             # git hook installation
│   ├── github/                  # GitHub API integration
│   │   ├── client.go            # GitHub API client (use go-gh for auth)
│   │   ├── pr.go                # create, update, get PRs
│   │   └── diagram.go           # stack diagram rendering for PR bodies
│   └── ui/                      # terminal output rendering
│       ├── graph.go             # git-log-style graph rendering
│       ├── colors.go            # lipgloss style definitions
│       └── format.go            # shared formatting utilities
├── go.mod
├── go.sum
├── main.go
└── README.md
```

---

## Key Implementation Details

### Git Layer — Hybrid go-git + exec

Use go-git (`github.com/go-git/go-git/v5`) for **read-only** operations:
- `CurrentBranch()` — resolve HEAD to branch name
- `BranchExists(name)` — check ref existence
- `BranchTip(name)` — get abbreviated commit hash for a branch
- `RevParse(ref)` — resolve any ref to a commit hash

Use `exec.Command("git", ...)` for **mutations**:
- `Checkout(branch, create)` — `git checkout [-b] <branch>`
- `RebaseOnto(newBase, oldBase, branch)` — `git rebase --onto <newBase> <oldBase> <branch>`
- `Push(branch, force)` — `git push [--force-with-lease] origin <branch>`
- `Fetch(remote)` — `git fetch <remote>`
- `DeleteBranch(name)` — `git branch -D <name>`
- `NeedsRebase(branch, parent)` — `git rev-list --count <branch>..<parent>` (returns true if count > 0)
- `RebaseContinue()` — `git rebase --continue`
- `RebaseAbort()` — `git rebase --abort`

All exec commands should capture both stdout and stderr. For user-facing mutations (rebase, push), stream output to the terminal in real-time so the user sees git's progress/conflict messages.

### Stack Store — Persistence

```go
// pkg/stack/store.go

type Store struct {
    State   *State
    gitDir  string  // path to .git directory
}

// LoadStore opens the repo, finds .git, and loads stacked.json
func LoadStore(repo *git.Repo) (*Store, error) { ... }

// Save writes state back to .git/stacked.json
func (s *Store) Save() error { ... }

// CurrentStack returns the active stack, or nil
func (s *Store) CurrentStack() *Stack { ... }

// StackForBranch finds which stack a branch belongs to (for when the user
// checks out a stack branch manually via git)
func (s *Store) StackForBranch(branchName string) (*Stack, string) { ... }
```

The store should detect the current stack not just from `current_stack` in the JSON, but also by checking if the current git branch belongs to any known stack. This handles the case where a user does a manual `git checkout` to a stack branch.

### Cascade Rebase

This is the most critical and error-prone operation. When any branch in the stack changes (or when the base branch moves forward), all branches above it must be rebased in order.

```
Algorithm:
1. Save the current branch name
2. For each branch in order (bottom to top):
   a. Determine its parent (previous branch, or base for the first)
   b. Run `git rebase --onto <parent> <parent> <branch>`
   c. If rebase fails (conflicts), save progress state and exit with instructions
3. Restore the original branch
```

**Conflict handling:** When a rebase hits conflicts mid-cascade, write a `.git/stacked-rebase-state.json` file that records:
- Which stack is being rebased
- Which branch index the conflict is on
- The original branch the user was on

`stacked continue` resumes: it runs `git rebase --continue`, and if successful, continues the cascade from the next branch. `stacked abort` runs `git rebase --abort` and cleans up the state file.

### GitHub Integration

Use `github.com/cli/go-gh/v2` for authentication. This piggybacks off `gh auth login`, so users don't need to configure tokens separately. If `go-gh` is not available, fall back to `GITHUB_TOKEN` environment variable.

**PR Creation (`stacked review`):**
- For each branch in the stack that doesn't have a PR yet:
  - Base branch = previous branch in stack (NOT main, except for the first branch)
  - Title = branch title from metadata
  - Body = stack diagram (see below) + any user content
- For branches that already have PRs, update the PR body with the latest stack diagram
- After creating PRs, save PR numbers back to `.git/stacked.json`

**PR Retargeting (`stacked sync`):**
- After pruning merged branches, the new bottom of the stack needs its PR base changed to `main` (or whatever the stack base is)
- Use the GitHub API `PATCH /repos/{owner}/{repo}/pulls/{pull_number}` with `{ "base": "main" }`

### Stack Diagram in PR Bodies

Each PR body gets a navigable stack diagram rendered in markdown, wrapped in HTML comment sentinels for idempotent updates.

**Format:**
```markdown
<!-- stacked:stack-diagram -->
> **Stack**:
> 3. [feat(svc): add logout](https://github.com/owner/repo/pull/125)
> 2. **feat(svc): add token refresh** ← *this PR*
> 1. [feat(svc): add user auth](https://github.com/owner/repo/pull/123)
>
> ⬇ Base: `main`
<!-- /stacked:stack-diagram -->
```

Rendering logic:
- The current PR's branch is **bolded** with a `← this PR` marker
- Other branches with PRs are rendered as clickable links
- Branches without PRs are rendered as plain text with `*(not yet opened)*`
- Stack is rendered top-down (tip of stack at top, base at bottom)
- The sentinel HTML comments allow find-and-replace on subsequent updates without touching user-written content in the PR body

**Update logic:** On every `stacked push` or `stacked review`, iterate all PRs in the stack, fetch their current body, replace the content between sentinels with the new diagram, and PATCH if changed. Skip the API call if the body hasn't changed (avoid unnecessary writes).

### CLI Graph Output (`stacked status`)

Render a git-log-style graph in the terminal using lipgloss for colors. Output format:

```
* abc1234 feat(svc): add logout  (HEAD -> stack/.../03-add-logout, PR #125)
|
* def5678 feat(svc): add token refresh  (stack/.../02-add-token-refresh, PR #124)
|
* 9ab0123 feat(svc): add user auth  (stack/.../01-add-user-auth, PR #123 ✓ merged)
|
○ main (base)
```

Symbols:
- `*` (yellow) — current branch (HEAD)
- `*` (cyan) — other stack branches
- `✓` (green) — merged
- `✗` (red) — conflict state
- `○` (dim) — base branch

Decorations in parentheses, git-style:
- `HEAD -> <branch>` in green for current branch
- Branch name in cyan for other branches
- PR number in dim, with status suffix: `✓ merged`, `⚠ needs rebase`, `✗ conflict`
- `no PR` in dim if no PR created yet

**Verbose mode (`-v`):** Include abbreviated commit hashes before the title.

**Needs-rebase detection:** For each branch, check if its parent has commits not in the branch via `git rev-list --count <branch>..<parent>`. If count > 0, mark as `⚠ needs rebase`.

### Auto-Sync Staleness Check

Every command should begin with a cheap check: compare `<base>` to `origin/<base>`. If they differ, the remote has moved (someone merged something). Print a hint:

```
hint: main has new commits. Run `stacked sync` to update your stack.
```

Don't auto-sync — just inform. The user should be in control of when rebases happen.

Implementation:
```go
func SyncHint(repo *git.Repo, s *stack.Stack) {
    local, _ := repo.RevParse(s.Base)
    remote, _ := repo.RevParse("origin/" + s.Base)
    if local != remote {
        fmt.Fprintf(os.Stderr, "hint: %s has new commits. Run `stacked sync` to update.\n", s.Base)
    }
}
```

### Post-Merge Sync (`stacked sync`)

Full algorithm:
1. `git fetch origin`
2. `git checkout <base> && git pull --ff-only origin <base>`
3. Walk branches from bottom of stack:
   - For each branch, check if its PR is merged (via GitHub API)
   - If merged, delete local branch (`git branch -D`), remove from stack metadata
   - If not merged, stop walking — remaining branches are the active stack
4. If any branches were pruned:
   - Cascade rebase remaining branches onto the (now-updated) base
   - Retarget the new bottom PR's base to `<base>` via GitHub API
   - Force-push all rebased branches
   - Update stack diagrams in all remaining PR bodies
5. Save updated stack metadata

### Git Hooks (Optional)

`stacked init --hooks` installs hooks that trigger sync hints:

**`post-merge`** (fires after `git pull`):
```bash
#!/bin/sh
stacked sync --if-needed --quiet 2>/dev/null
```

**`post-checkout`** (fires on branch switch):
```bash
#!/bin/sh
if [ "$3" = "1" ]; then
    stacked sync --if-needed --quiet 2>/dev/null
fi
```

Hook installation should be additive — check for existing hooks and append rather than overwrite. Include a `# stacked` marker comment so the tool can detect if hooks are already installed.

---

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/go-git/go-git/v5` | Git reads (branch resolution, ref checks) |
| `github.com/cli/go-gh/v2` | GitHub API auth (piggyback off `gh auth`) |
| `github.com/charmbracelet/lipgloss` | Terminal styling |
| `github.com/charmbracelet/tree` | Tree/graph rendering (optional, may be easier to custom-render) |

---

## Error Handling Principles

- Every git exec command must check for errors and return them with context (`fmt.Errorf("failed to checkout %s: %w", branch, err)`)
- Rebase conflicts are expected, not exceptional. Print clear instructions: "Resolve conflicts, then run `stacked continue`"
- GitHub API failures should not corrupt local state. Save stack metadata *after* successful git operations, not before.
- If `.git/stacked.json` is missing or corrupt, commands should fail with a clear message: "No stacked configuration found. Run `stacked new` to create a stack."

---

## Testing Strategy

- **Unit tests:** Stack data model operations (Parent, Descendants, IndexOf, slug generation)
- **Integration tests:** Use `git init` in a temp directory, create commits, and exercise the full command flow. Shell out to real `git` — don't mock it.
- **GitHub tests:** Mock the GitHub API client interface. Define a `GitHubClient` interface with `CreatePR`, `UpdatePR`, `GetPR`, `UpdatePRBase` methods and provide a mock implementation for tests.

---

## Out of Scope (v1)

- Multi-forge support (GitLab, Bitbucket) — GitHub only for v1
- Interactive TUI (bubbletea) — plain CLI output first
- Background daemon for auto-sync
- Parallel stack support (diamond-shaped dependency graphs)
- Commit-per-PR enforcement (user manages their own commits)