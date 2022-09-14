## ko create

Create the input files with image references resolved to built/pushed image digests.

### Synopsis

This sub-command finds import path references within the provided files, builds them into Go binaries, containerizes them, publishes them, and then feeds the resulting yaml into "kubectl create".

```
ko create -f FILENAME [flags]
```

### Examples

```

  # Build and publish import path references to a Docker
  # Registry as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # Then, feed the resulting yaml into "kubectl create".
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local was passed.
  ko create -f config/

  # Build and publish import path references to a Docker
  # Registry preserving import path names as:
  #   ${KO_DOCKER_REPO}/<import path>
  # Then, feed the resulting yaml into "kubectl create".
  ko create --preserve-import-paths -f config/

  # Build and publish import path references to a Docker
  # daemon as:
  #   ko.local/<import path>
  # Then, feed the resulting yaml into "kubectl create".
  ko create --local -f config/

  # Create from stdin:
  cat config.yaml | ko create -f -

  # Any flags passed after '--' are passed to 'kubectl apply' directly:
  ko apply -f config -- --namespace=foo --kubeconfig=cfg.yaml

```

### Options

```
      --bare                     Whether to just use KO_DOCKER_REPO without additional context (may not work properly with --tags).
  -B, --base-import-paths        Whether to use the base path without MD5 hash after KO_DOCKER_REPO (may not work properly with --tags).
      --disable-optimizations    Disable optimizations when building Go code. Useful when you want to interactively debug the created container.
  -f, --filename strings         Filename, directory, or URL to files to use to create the resource
  -h, --help                     help for create
      --image-label strings      Which labels (key=value) to add to the image.
      --image-refs string        Path to file where a list of the published image references will be written.
      --insecure-registry        Whether to skip TLS verification on the registry
  -j, --jobs int                 The maximum number of concurrent builds (default GOMAXPROCS)
  -L, --local                    Load into images to local docker daemon.
      --oci-layout-path string   Path to save the OCI image layout of the built images
      --platform strings         Which platform to use when pulling a multi-platform base. Format: all | <os>[/<arch>[/<variant>]][,platform]*
  -P, --preserve-import-paths    Whether to preserve the full import path after KO_DOCKER_REPO.
      --push                     Push images to KO_DOCKER_REPO (default true)
  -R, --recursive                Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.
      --sbom string              The SBOM media type to use (none will disable SBOM synthesis and upload, also supports: spdx, cyclonedx, go.version-m). (default "spdx")
      --sbom-dir string          Path to file where the SBOM will be written.
  -l, --selector string          Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
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

