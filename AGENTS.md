# AGENTS.md — Instructions for AI Coding Agents

## Project

`stacked` is a Go CLI for managing stacked pull requests. Read `stacked-prd.md` for full project requirements, data model, and architecture before writing any code.

## Language & Tooling

- **Go 1.22+** — use modern idioms (range-over-int, structured logging via `log/slog`)
- **Module path:** `github.com/kerb/stacked` (replace with actual module path)
- **Build:** `go build -o stacked ./cmd/stacked`
- **Test:** `go test ./...`
- **Lint:** `golangci-lint run` — zero warnings policy
- **Format:** `gofmt` / `goimports` — all code must be formatted before committing

## Code Style

### General

- No `init()` functions. Ever.
- No global mutable state. Pass dependencies explicitly.
- Errors are values. Wrap with context: `fmt.Errorf("checkout %s: %w", branch, err)`. Never discard errors silently — if you intentionally ignore one, assign to `_` with a comment explaining why.
- No `panic` or `os.Exit` outside of `main.go`. Return errors up the call stack.
- Prefer early returns over deep nesting. Guard clauses first.
- Keep functions under 50 lines. If it's longer, extract a helper.
- No `interface{}` or `any` unless absolutely necessary. Use generics or concrete types.

### Naming

- Package names: short, lowercase, single word. `stack`, `git`, `ui` — not `stackmanager`, `githelper`.
- Exported types: `Stack`, `Branch`, `Store` — noun, no prefix/suffix.
- Constructors: `NewStore(...)`, `Open(...)` — not `CreateStore`.
- Methods that return bool: `IsEmpty()`, `NeedsRebase()`, `HasPR()`.
- CLI commands: register in their own file, one file per command. `new.go`, `next.go`, `push.go`.
- Test files: `*_test.go` in the same package. Use table-driven tests.

### Error Messages

- Lowercase, no trailing punctuation: `"failed to push branch"` not `"Failed to push branch."`
- Include the thing that failed: `"checkout %s: %w"` not just `"checkout failed"`
- User-facing errors should suggest next steps: `"resolve conflicts, then run 'stacked continue'"`

### Comments

- No obvious comments. `// Create a new stack` above `func NewStack()` is noise.
- Comment *why*, not *what*. Explain non-obvious decisions, gotchas, workarounds.
- Package-level doc comments on every package.
- Exported functions get doc comments. Internal helpers usually don't need them unless the logic is subtle.

## Architecture Rules

### Package Boundaries

```
cmd/        → Cobra command definitions. Thin layer: parse flags, call stack/git/github, format output.
pkg/stack/  → Pure data model. NO git operations, NO GitHub API calls, NO terminal output.
pkg/git/    → All git interaction. go-git for reads, exec for mutations. No stack awareness.
pkg/github/ → GitHub API only. No git operations. No terminal output.
pkg/ui/     → All terminal formatting and styling. lipgloss styles, graph rendering. No business logic.
```

**Dependency direction:** `cli → stack, git, github, ui`. Never `stack → cli` or `git → stack`. The `stack` package is a pure data library with zero external dependencies beyond stdlib.

### Git Layer

- **Read operations** use go-git (`github.com/go-git/go-git/v5`): branch resolution, ref lookup, commit hash retrieval. These are fast and avoid subprocess overhead.
- **Write operations** shell out to `git` via `exec.Command`: checkout, rebase, push, fetch, branch deletion. This is intentional — go-git has no rebase support and its write path is unreliable for our use cases.
- All `exec.Command` calls must set `cmd.Dir` to the repo root path.
- For user-facing mutations (rebase, push), pipe stdout/stderr to `os.Stdout`/`os.Stderr` so the user sees real-time git output.
- For internal checks (rev-list, rev-parse), capture output into a string.

### State Management

- Stack metadata lives in `.git/stacked.json`. Load once at command start, save once at command end.
- **Never save state before git operations complete.** If a rebase fails, the state file should still reflect the pre-rebase reality.
- Rebase-in-progress state goes in `.git/stacked-rebase-state.json` — separate from the main state file so a failed rebase doesn't corrupt the stack.

### GitHub API

- Use `github.com/cli/go-gh/v2` for auth. Fall back to `GITHUB_TOKEN` env var.
- Define a `GitHubClient` interface in `pkg/github/client.go`:
  ```go
  type Client interface {
      CreatePR(ctx context.Context, input CreatePRInput) (*PR, error)
      UpdatePR(ctx context.Context, owner, repo string, number int, input UpdatePRInput) error
      GetPR(ctx context.Context, owner, repo string, number int) (*PR, error)
      UpdatePRBase(ctx context.Context, owner, repo string, number int, base string) error
  }
  ```
- Tests use a mock implementation of this interface. Never call the real GitHub API in tests.

## Testing

### Unit Tests

- Table-driven tests for all pure functions (slug generation, stack operations, diagram rendering).
- Use `t.Run(name, ...)` for subtests.
- Test file goes next to the source file: `slug.go` → `slug_test.go`.

### Integration Tests

- Create a real git repo in `t.TempDir()`. Make real commits. Run real git commands.
- Use build tags if integration tests are slow: `//go:build integration`
- Test the full command flow: `new → next → next → status → push`.
- Do NOT mock git. The whole point is to verify real git behavior.

### What to Test

- Stack data model: Parent(), Descendants(), IndexOf(), Tip()
- Slug generation: edge cases (special characters, very long names, unicode)
- Graph rendering: verify output string matches expected format
- Diagram rendering: verify markdown output, sentinel tags, idempotent updates
- Store: load/save round-trip, corrupt file handling, missing file handling
- Rebase cascade: multi-branch rebase, conflict detection
- Sync: merged branch pruning, retargeting

### What NOT to Test

- Cobra command wiring (trust the framework)
- lipgloss color output (visual, not unit-testable)
- GitHub API response parsing (trust go-gh)

## CLI Output Conventions

- **Normal output** goes to stdout.
- **Hints and warnings** go to stderr (so piping works: `stacked status | grep merged`).
- **Progress messages** during multi-step operations use a consistent prefix:
  ```
  pushing stack/feat/01-auth...
  pushing stack/feat/02-tokens...
  updating PR #123 diagram...
  ```
- **Errors** are printed by the cobra `RunE` handler, not inside business logic. Business logic returns errors, CLI layer prints them.
- No spinners or progress bars in v1. Keep it simple.

## Common Pitfalls to Avoid

1. **Don't rebase without saving the current branch first.** `git rebase --onto` changes HEAD. Always save and restore.
2. **Don't assume branch order matches stack order.** Always use the stack metadata as the source of truth, not git reflog or alphabetical sorting.
3. **Don't update PR body if nothing changed.** Compare old vs new body before making API calls. GitHub rate limits are real.
4. **Don't delete branches that have open PRs targeting them.** Sync must retarget first, then delete.
5. **Don't use `git push --force`.** Always `--force-with-lease` to avoid overwriting someone else's work.
6. **Don't shell out to `gh` CLI.** Use the Go library `go-gh` directly. Shelling out to `gh` adds a runtime dependency and is slower.
7. **Don't put business logic in the CLI layer.** `cmd/*.go` should be thin wrappers: parse args, call functions, format output. If a CLI handler is over 30 lines, you're doing too much there.

## Implementation Order

Build and test each phase before moving to the next.

### Phase 1 — Core Stack Operations

Files: `pkg/stack/`, `pkg/git/`, `cmd/new.go`, `cmd/next.go`, `cmd/status.go`, `pkg/ui/`

1. `pkg/stack/stack.go` — types and methods
2. `pkg/stack/store.go` — JSON persistence
3. `pkg/stack/slug.go` — branch name generation
4. `pkg/git/repo.go` — go-git read operations
5. `pkg/git/exec.go` — git exec wrapper
6. `pkg/ui/colors.go` — lipgloss style definitions
7. `pkg/ui/graph.go` — git-log-style graph renderer
8. `cmd/new.go` — create new stack
9. `cmd/next.go` — add branch to stack
10. `cmd/status.go` — render stack graph
11. `cmd/navigate.go` — up/down/goto
12. `cmd/stacked/main.go` — wire everything together

**Checkpoint:** You should be able to run `stacked new`, `stacked next`, `stacked status`, `stacked up/down` and see a working graph.

### Phase 2 — Rebase & Push

Files: `cmd/rebase.go`, `cmd/push.go`, `cmd/continue_abort.go`

1. Cascade rebase logic in `pkg/git/exec.go`
2. Rebase state tracking (`.git/stacked-rebase-state.json`)
3. `stacked rebase` command
4. `stacked continue` / `stacked abort`
5. `stacked push` — force-with-lease all branches
6. Staleness check (compare local base to origin/base)

**Checkpoint:** Full local workflow works — create stack, make commits, rebase, push.

### Phase 3 — GitHub Integration

Files: `pkg/github/`, `cmd/review.go`, `cmd/sync.go`

1. `pkg/github/client.go` — GitHub API client with interface
2. `pkg/github/pr.go` — PR CRUD operations
3. `pkg/github/diagram.go` — stack diagram markdown rendering
4. `stacked review` — create/update PRs with correct base branches and diagrams
5. `stacked sync` — detect merged PRs, prune, retarget, rebase, update diagrams
6. `stacked list` — list all stacks

**Checkpoint:** End-to-end: create stack → push → create PRs → merge bottom PR on GitHub → sync → see updated stack.

### Phase 4 — Polish

1. `stacked init --hooks` — git hook installation
2. Better error messages and edge case handling
3. `--quiet` and `--verbose` global flags
4. README.md with usage examples
5. `goreleaser` config for binary releases