package stack

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()

	store := &Store{
		gitDir: dir,
		State: &State{
			Version: 1,
			Stacks: map[string]*Stack{
				"feat-auth": {
					Base: "main",
					Branches: []Branch{
						{Name: "stack/feat-auth/01-auth", Title: "add auth", CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
						{Name: "stack/feat-auth/02-tokens", Title: "add tokens", CreatedAt: time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)},
					},
				},
			},
			CurrentStack: "feat-auth",
		},
	}

	if err := store.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadStore(dir)
	if err != nil {
		t.Fatalf("LoadStore: %v", err)
	}

	if loaded.State.Version != 1 {
		t.Errorf("Version = %d, want 1", loaded.State.Version)
	}
	if loaded.State.CurrentStack != "feat-auth" {
		t.Errorf("CurrentStack = %q, want %q", loaded.State.CurrentStack, "feat-auth")
	}

	st := loaded.CurrentStack()
	if st == nil {
		t.Fatal("CurrentStack() returned nil")
	}
	if len(st.Branches) != 2 {
		t.Errorf("got %d branches, want 2", len(st.Branches))
	}
	if st.Branches[0].Title != "add auth" {
		t.Errorf("first branch title = %q, want %q", st.Branches[0].Title, "add auth")
	}
}

func TestStoreLoadMissing(t *testing.T) {
	dir := t.TempDir()

	store, err := LoadStore(dir)
	if err != nil {
		t.Fatalf("LoadStore on missing file: %v", err)
	}
	if store.State.Version != 1 {
		t.Errorf("Version = %d, want 1", store.State.Version)
	}
	if len(store.State.Stacks) != 0 {
		t.Errorf("Stacks should be empty, got %d", len(store.State.Stacks))
	}
}

func TestStoreLoadCorrupt(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, stateFile), []byte("{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadStore(dir)
	if err == nil {
		t.Fatal("LoadStore on corrupt file should error")
	}
}

func TestStackForBranch(t *testing.T) {
	store := &Store{
		gitDir: t.TempDir(),
		State: &State{
			Version: 1,
			Stacks: map[string]*Stack{
				"feat-a": {Base: "main", Branches: []Branch{{Name: "stack/feat-a/01-a", Title: "a"}}},
				"feat-b": {Base: "main", Branches: []Branch{{Name: "stack/feat-b/01-b", Title: "b"}}},
			},
		},
	}

	st, name := store.StackForBranch("stack/feat-b/01-b")
	if name != "feat-b" {
		t.Errorf("StackForBranch name = %q, want %q", name, "feat-b")
	}
	if st == nil {
		t.Fatal("StackForBranch returned nil stack")
	}

	st, name = store.StackForBranch("unknown-branch")
	if st != nil || name != "" {
		t.Error("StackForBranch should return nil for unknown branch")
	}
}

func TestResolveCurrentStack(t *testing.T) {
	store := &Store{
		gitDir: t.TempDir(),
		State: &State{
			Version: 1,
			Stacks: map[string]*Stack{
				"feat-a": {Base: "main", Branches: []Branch{{Name: "stack/feat-a/01-a", Title: "a"}}},
				"feat-b": {Base: "main", Branches: []Branch{{Name: "stack/feat-b/01-b", Title: "b"}}},
			},
			CurrentStack: "feat-a",
		},
	}

	// Current stack matches branch
	st, name := store.ResolveCurrentStack("stack/feat-a/01-a")
	if name != "feat-a" || st == nil {
		t.Error("should resolve from current_stack")
	}

	// Branch belongs to different stack — should switch
	st, name = store.ResolveCurrentStack("stack/feat-b/01-b")
	if name != "feat-b" {
		t.Errorf("should resolve to feat-b, got %q", name)
	}
	if store.State.CurrentStack != "feat-b" {
		t.Error("should update current_stack")
	}

	// Unknown branch
	st, name = store.ResolveCurrentStack("unknown")
	if st != nil || name != "" {
		t.Error("should return nil for unknown branch")
	}
}
