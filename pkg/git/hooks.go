package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const hookMarker = "# stacked"

var hooks = map[string]string{
	"post-merge": `#!/bin/sh
# stacked
stacked sync --if-needed --quiet 2>/dev/null
`,
	"post-checkout": `#!/bin/sh
# stacked
if [ "$3" = "1" ]; then
    stacked sync --if-needed --quiet 2>/dev/null
fi
`,
}

// InstallHooks installs stacked git hooks. If hooks already exist,
// it appends the stacked section (unless already installed).
func (r *Repo) InstallHooks() error {
	hooksDir := filepath.Join(r.GitDir(), "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return fmt.Errorf("create hooks dir: %w", err)
	}

	for name, content := range hooks {
		hookPath := filepath.Join(hooksDir, name)
		if err := installHook(hookPath, content); err != nil {
			return fmt.Errorf("install %s hook: %w", name, err)
		}
	}

	return nil
}

// HasHooks reports whether stacked hooks are already installed.
func (r *Repo) HasHooks() bool {
	hooksDir := filepath.Join(r.GitDir(), "hooks")
	for name := range hooks {
		data, err := os.ReadFile(filepath.Join(hooksDir, name))
		if err != nil {
			return false
		}
		if !strings.Contains(string(data), hookMarker) {
			return false
		}
	}
	return true
}

func installHook(path, content string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if strings.Contains(string(existing), hookMarker) {
		return nil // already installed
	}

	if len(existing) > 0 {
		// Append to existing hook
		content = string(existing) + "\n" + content
	}

	return os.WriteFile(path, []byte(content), 0o755)
}
