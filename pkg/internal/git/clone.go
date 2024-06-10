package git

import (
	"context"
	"fmt"
	"os/exec"
)

// Clone the git repository from the repoURL to the specified dir.
func Clone(ctx context.Context, dir string, repoURL string) error {
	rc := runConfig{
		dir:  dir,
		args: []string{"clone", "--depth", "1", repoURL},
	}

	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, dir)
	cmd.Dir = dir

	_, err := run(ctx, rc)
	if err != nil {
		return fmt.Errorf("running git clone: %v", err)
	}

	return nil
}
