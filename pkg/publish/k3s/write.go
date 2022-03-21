package k3s

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	limaInstanceEnvKey            = "LIMA_INSTANCE"
	rancherDesktopLimaInstanceKey = "0"
)

// Tag adds a tag to an already existent image.
func Tag(ctx context.Context, src, dest name.Tag) error {
	li, ok := os.LookupEnv(limaInstanceEnvKey)
	if !ok {
		li = rancherDesktopLimaInstanceKey
	}
	env := buildCommandEnv(li)
	ctl, err := findNerdctl(li)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, ctl, "--namespace=k8s.io", "tag", src.String(), dest.String())
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to tag image to instance %q: %w", li, err)
	}

	return nil
}

// Write saves the image into the k3s nodes as the given tag.
func Write(ctx context.Context, tag name.Tag, img v1.Image) error {
	pr, pw := io.Pipe()

	grp := errgroup.Group{}
	grp.Go(func() error {
		return pw.CloseWithError(tarball.Write(tag, img, pw))
	})

	li, ok := os.LookupEnv(limaInstanceEnvKey)
	//TODO(kamesh) for now its assumed that if no LIMA_INSTANCE env is defined it defaults to Rancher Desktop
	// is this safe assumption or need to find other ways??
	if !ok {
		li = rancherDesktopLimaInstanceKey
	}
	env := buildCommandEnv(li)

	var stdErr bytes.Buffer
	//check of nerdctl exists on the system
	ctl, err := findNerdctl(li)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, ctl, "--namespace=k8s.io", "load")
	cmd.Stdin = pr
	cmd.Env = env
	cmd.Stderr = &stdErr
	if err := cmd.Run(); err != nil {
		log.Printf("%s", stdErr.String())
		return fmt.Errorf("failed to load image to instance %q: %w", li, err)
	}

	if err := grp.Wait(); err != nil {
		return fmt.Errorf("failed to write intermediate tarball representation: %w", err)
	}

	return nil
}

//buildCommandEnv adds the required environment variables that will be passed to the
// command context
//TODO(kamesh) add other required environment variables
func buildCommandEnv(li string) []string {
	var env = make([]string, 5)

	env[0] = fmt.Sprintf("HOME=%s", os.Getenv("HOME"))
	env[1] = fmt.Sprintf("LIMA_INSTANCE=%s", li)
	env[2] = fmt.Sprintf("PATH=%s", os.Getenv("PATH"))

	return env
}

//findNerdctl helps to find the nerdctl to use
//TODO(kamesh) improve the nerdctl find
//TODO(kamesh) not very efficient
func findNerdctl(li string) (string, error) {
	var nerdctlPath string
	// use rancher desktop nerdctl wrapper script
	if li == "0" {
		f, err := exec.LookPath("nerdctl")
		if err != nil {
			return "", err
		}
		nerdctlPath, err = filepath.Abs(f)
		if err != nil {
			return "", err
		}
	} else {
		//if rancher desktop is on the system there should be an alternate script nerdctl.lima
		f, err1 := exec.LookPath("nerdctl.lima")
		nerdctlPath, err1 = filepath.Abs(f)
		if err1 != nil {
			return "", err1
		}
	}

	_, err := os.Stat(nerdctlPath)
	if err != nil {
		return "", err
	}

	return nerdctlPath, nil
}
