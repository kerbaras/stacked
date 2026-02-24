package stack

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "add user auth", "add-user-auth"},
		{"with scope", "feat(svc): add user auth", "feat-svc-add-user-auth"},
		{"uppercase", "Add User Auth", "add-user-auth"},
		{"special chars", "fix: bug #123 (urgent!)", "fix-bug-123-urgent"},
		{"unicode stripped", "feat: añadir café", "feat-aadir-caf"},
		{"consecutive special", "feat:::add---thing", "feat-add-thing"},
		{"leading trailing dashes", "---hello---", "hello"},
		{"empty string", "", ""},
		{
			"long name truncated",
			"this is a very long title that should be truncated at fifty characters exactly right here",
			"this-is-a-very-long-title-that-should-be-truncated",
		},
		{"numbers preserved", "step 1 of 3", "step-1-of-3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStackName(t *testing.T) {
	got := StackName("feat(svc): add user auth")
	want := "feat-svc-add-user-auth"
	if got != want {
		t.Errorf("StackName() = %q, want %q", got, want)
	}
}

func TestBranchName(t *testing.T) {
	const defaultTpl = "stack/%name/%02d-%slug"

	tests := []struct {
		name      string
		template  string
		stackName string
		index     int
		title     string
		want      string
	}{
		{
			"first branch",
			defaultTpl,
			"feat-svc-add-user-auth", 1, "add user auth",
			"stack/feat-svc-add-user-auth/01-add-user-auth",
		},
		{
			"tenth branch",
			defaultTpl,
			"feat-svc-add-user-auth", 10, "final cleanup",
			"stack/feat-svc-add-user-auth/10-final-cleanup",
		},
		{
			"custom template",
			"refs/%name/%03d/%slug",
			"my-stack", 3, "add auth",
			"refs/my-stack/003/add-auth",
		},
		{
			"no index",
			"branch/%name/%slug",
			"my-stack", 5, "setup",
			"branch/my-stack/setup",
		},
		{
			"plain index",
			"stack/%name/%d-%slug",
			"my-stack", 7, "cleanup",
			"stack/my-stack/7-cleanup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BranchName(tt.template, tt.stackName, tt.index, tt.title)
			if got != tt.want {
				t.Errorf("BranchName() = %q, want %q", got, tt.want)
			}
		})
	}
}
