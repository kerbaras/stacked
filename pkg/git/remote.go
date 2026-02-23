package git

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	sshRemoteRe   = regexp.MustCompile(`^git@[^:]+:([^/]+)/([^/.]+?)(?:\.git)?$`)
	httpsRemoteRe = regexp.MustCompile(`^https?://[^/]+/([^/]+)/([^/.]+?)(?:\.git)?$`)
)

// RemoteOwnerRepo extracts the owner and repo from the origin remote URL.
func (r *Repo) RemoteOwnerRepo() (string, string, error) {
	out, err := r.execCapture("remote", "get-url", "origin")
	if err != nil {
		return "", "", fmt.Errorf("get origin URL: %w", err)
	}
	url := strings.TrimSpace(out)

	if m := sshRemoteRe.FindStringSubmatch(url); m != nil {
		return m[1], m[2], nil
	}
	if m := httpsRemoteRe.FindStringSubmatch(url); m != nil {
		return m[1], m[2], nil
	}

	return "", "", fmt.Errorf("cannot parse remote URL: %s", url)
}
