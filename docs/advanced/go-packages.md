# Go Packages

`ko`'s functionality can be consumed as a library in a Go application.

To build an image, use [`pkg/build`](https://pkg.go.dev/github.com/google/ko/pkg/build), and publish it with [`pkg/publish`](https://pkg.go.dev/github.com/google/ko/pkg/publish).

This is a minimal example of using the packages together, to implement the core subset of `ko`'s functionality:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/publish"
)

const (
	baseImage  = "cgr.dev/chainguard/static:latest"
	targetRepo = "example.registry/my-repo"
	importpath = "github.com/my-org/miniko"
	commitSHA  = "deadbeef"
)

func main() {
	ctx := context.Background()

	b, err := build.NewGo(ctx, ".",
		build.WithPlatforms("linux/amd64"), // only build for these platforms.
		build.WithBaseImages(func(ctx context.Context, _ string) (name.Reference, build.Result, error) {
			ref := name.MustParseReference(baseImage)
			base, err := remote.Index(ref, remote.WithContext(ctx))
			return ref, base, err
		}))
	if err != nil {
		log.Fatalf("NewGo: %v", err)
	}
	r, err := b.Build(ctx, importpath)
	if err != nil {
		log.Fatalf("Build: %v", err)
	}

	p, err := publish.NewDefault(targetRepo,                 // publish to example.registry/my-repo
		publish.WithTags([]string{commitSHA}),               // tag with :deadbeef
		publish.WithAuthFromKeychain(authn.DefaultKeychain)) // use credentials from ~/.docker/config.json
	if err != nil {
		log.Fatalf("NewDefault: %v", err)
	}
	ref, err := p.Publish(ctx, r, importpath)
	if err != nil {
		log.Fatalf("Publish: %v", err)
	}
	fmt.Println(ref.String())
}
```
