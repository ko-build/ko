## ko run

A variant of `kubectl run` that containerizes IMPORTPATH first.

### Synopsis

This sub-command combines "ko build" and "kubectl run" to support containerizing and running Go binaries on Kubernetes in a single command.

```
ko run IMPORTPATH [flags]
```

### Examples

```

  # Publish the image and run it on Kubernetes as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local and --preserve-import-paths were passed.
  ko run github.com/foo/bar/cmd/baz

  # This supports relative import paths as well.
  ko run ./cmd/baz

  # You can also supply args and flags to the command.
  ko run ./cmd/baz -- -v arg1 arg2 --yes
```

### Options

```
      --bare[=true|false]                    Whether to just use KO_DOCKER_REPO without additional context (may not work properly with --tags).
  -B, --base-import-paths[=true|false]       Whether to use the base path without MD5 hash after KO_DOCKER_REPO (may not work properly with --tags).
      --debug[=true|false]                   Include Delve debugger into image and wrap around ko-app. This debugger will listen to port 40000.
      --disable-optimizations[=true|false]   Disable optimizations when building Go code. Useful when you want to interactively debug the created container.
  -h, --help[=true|false]                    help for run
      --image-annotation strings             Which annotations (key=value[,key=value]) to add to the OCI manifest.
      --image-label strings                  Which labels (key=value[,key=value]) to add to the image.
      --image-refs string                    Path to file where a list of the published image references will be written.
      --image-user string                    The default user the image should be run as.
      --insecure-registry[=true|false]       Whether to skip TLS verification on the registry
  -j, --jobs int                             The maximum number of concurrent builds (default GOMAXPROCS)
      --ldflags strings                      ldflags to pass to go build (may be repeated)
  -L, --local[=true|false]                   Load into images to local docker daemon.
      --local-podman[=true|false]            Load images into local Podman storage instead of a registry.
      --oci-layout-path string               Path to save the OCI image layout of the built images
      --platform strings                     Which platform to use when pulling a multi-platform base. Format: all | <os>[/<arch>[/<variant>]][,platform]*
  -P, --preserve-import-paths[=true|false]   Whether to preserve the full import path after KO_DOCKER_REPO.
      --push[=true|false]                    Push images to KO_DOCKER_REPO (default true)
      --sbom string                          The SBOM media type to use (none will disable SBOM synthesis and upload). (default "spdx")
      --sbom-dir string                      Path to directory where the SBOM will be written.
      --tag-only[=true|false]                Include tags but not digests in resolved image references. Useful when digests are not preserved when images are repopulated.
  -t, --tags strings                         Which tags to use for the produced image instead of the default 'latest' tag (may not work properly with --base-import-paths or --bare). (default [latest])
      --tarball string                       File to save images tarballs
```

### Options inherited from parent commands

```
  -v, --verbose[=true|false]   Enable debug logs
```

### SEE ALSO

* [ko](ko.md)	 - Rapidly iterate with Go, Containers, and Kubernetes.

