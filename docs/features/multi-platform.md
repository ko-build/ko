# Multi-Platform Images

Because Go supports cross-compilation to other CPU architectures and operating systems, `ko` excels at producing multi-platform images.

To build and push an image for all platforms supported by the configured base image, simply add `--platform=all`.
This will instruct `ko` to look up all the supported platforms in the base image, execute `GOOS=<os> GOARCH=<arch> GOARM=<variant> go build` for each platform, and produce a manifest list containing an image for each platform.

You can also select specific platforms, for example, `--platform=linux/amd64,linux/arm64`.

`ko` also has experimental support for building for Windows images.
See [FAQ](../../advanced/faq#can-i-build-windows-containers).

