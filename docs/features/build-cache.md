# Build Cache

Because `ko` just runs `go build` in your normal development environment, it automatically reuses your [`go build` cache](https://pkg.go.dev/cmd/go#hdr-Build_and_test_caching) from previous builds, making iterative development faster.

`ko` also avoids pushing blobs to the remote image registry if they're already present, making pushes faster.

You can make `ko` even faster by setting the `KOCACHE` environment variable.
This tells `ko` to store a local mapping between the `go build` inputs to the image layer that they produce, so `go build` can be skipped entirely if the layer is already present in the image registry.

