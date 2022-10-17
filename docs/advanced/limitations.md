# Limitations

`ko` works best when your application has no dependencies on the underlying image.

This means `ko` is ideal when you don't require [cgo](https://pkg.go.dev/cmd/cgo), and builds are executed with `CGO_ENABLED=0` by default.

To install other OS packages, make those available in your [configured base image](../../configuration).

`ko` only supports Go applications.
For a similar tool targeting Java applications, try [Jib](https://github.com/GoogleContainerTools/jib).
For other languages, try [apko](https://github.com/chainguard-dev/apko) and [melange](https://github.com/chainguard-dev/melange).

