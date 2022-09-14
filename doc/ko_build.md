## ko build

Build and publish container images from the given importpaths.

### Synopsis

This sub-command builds the provided import paths into Go binaries, containerizes them, and publishes them.

```
ko build IMPORTPATH... [flags]
```

### Examples

```

  # Build and publish import path references to a Docker Registry as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if --local and
  # --preserve-import-paths were passed.
  # If the import path is not provided, the current working directory is the
  # default.
  ko build github.com/foo/bar/cmd/baz github.com/foo/bar/cmd/blah

  # Build and publish a relative import path as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if --local and
  # --preserve-import-paths were passed.
  ko build ./cmd/blah

  # Build and publish a relative import path as:
  #   ${KO_DOCKER_REPO}/<import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if --local was passed.
  ko build --preserve-import-paths ./cmd/blah

  # Build and publish import path references to a Docker daemon as:
  #   ko.local/<import path>
  # This always preserves import paths.
  ko build --local github.com/foo/bar/cmd/baz github.com/foo/bar/cmd/blah
```

### Options

```
      --bare                     Whether to just use KO_DOCKER_REPO without additional context (may not work properly with --tags).
  -B, --base-import-paths        Whether to use the base path without MD5 hash after KO_DOCKER_REPO (may not work properly with --tags).
      --disable-optimizations    Disable optimizations when building Go code. Useful when you want to interactively debug the created container.
  -h, --help                     help for build
      --image-label strings      Which labels (key=value) to add to the image.
      --image-refs string        Path to file where a list of the published image references will be written.
      --insecure-registry        Whether to skip TLS verification on the registry
  -j, --jobs int                 The maximum number of concurrent builds (default GOMAXPROCS)
  -L, --local                    Load into images to local docker daemon.
      --oci-layout-path string   Path to save the OCI image layout of the built images
      --platform strings         Which platform to use when pulling a multi-platform base. Format: all | <os>[/<arch>[/<variant>]][,platform]*
  -P, --preserve-import-paths    Whether to preserve the full import path after KO_DOCKER_REPO.
      --push                     Push images to KO_DOCKER_REPO (default true)
      --sbom string              The SBOM media type to use (none will disable SBOM synthesis and upload, also supports: spdx, cyclonedx, go.version-m). (default "spdx")
      --sbom-dir string          Path to file where the SBOM will be written.
      --tag-only                 Include tags but not digests in resolved image references. Useful when digests are not preserved when images are repopulated.
  -t, --tags strings             Which tags to use for the produced image instead of the default 'latest' tag (may not work properly with --base-import-paths or --bare). (default [latest])
      --tarball string           File to save images tarballs
```

### Options inherited from parent commands

```
  -v, --verbose   Enable debug logs
```

### SEE ALSO

* [ko](ko.md)	 - Rapidly iterate with Go, Containers, and Kubernetes.

