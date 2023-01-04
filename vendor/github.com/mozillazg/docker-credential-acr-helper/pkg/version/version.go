package version

import (
	"fmt"
	"runtime"
)

var (
	ProjectName = "docker-credential-acr-helper"
	Version     = "0.0.0"
	GitCommit   = "unknown"
	Timestamp   = "unknown"
)

func UserAgent() string {
	return fmt.Sprintf("%s/%s (%s/%s) %s/%s",
		ProjectName, Version, runtime.GOOS, runtime.GOARCH, GitCommit, Timestamp)
}
