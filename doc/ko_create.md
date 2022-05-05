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
      --as string                      Username to impersonate for the operation (DEPRECATED)
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups. (DEPRECATED)
      --bare                           Whether to just use KO_DOCKER_REPO without additional context (may not work properly with --tags).
  -B, --base-import-paths              Whether to use the base path without MD5 hash after KO_DOCKER_REPO (may not work properly with --tags).
      --cache-dir string               Default cache directory (DEPRECATED)
      --certificate-authority string   Path to a cert file for the certificate authority (DEPRECATED)
      --client-certificate string      Path to a client certificate file for TLS (DEPRECATED)
      --client-key string              Path to a client key file for TLS (DEPRECATED)
      --cluster string                 The name of the kubeconfig cluster to use (DEPRECATED)
      --context string                 The name of the kubeconfig context to use (DEPRECATED)
      --disable-optimizations          Disable optimizations when building Go code. Useful when you want to interactively debug the created container.
  -f, --filename strings               Filename, directory, or URL to files to use to create the resource
  -h, --help                           help for create
      --image-label strings            Which labels (key=value) to add to the image.
      --image-refs string              Path to file where a list of the published image references will be written.
      --insecure-registry              Whether to skip TLS verification on the registry
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure (DEPRECATED)
  -j, --jobs int                       The maximum number of concurrent builds (default GOMAXPROCS)
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests. (DEPRECATED)
  -L, --local                          Load into images to local docker daemon.
  -n, --namespace string               If present, the namespace scope for this CLI request (DEPRECATED)
      --oci-layout-path string         Path to save the OCI image layout of the built images
      --password string                Password for basic authentication to the API server (DEPRECATED)
      --platform strings               Which platform to use when pulling a multi-platform base. Format: all | <os>[/<arch>[/<variant>]][,platform]*
  -P, --preserve-import-paths          Whether to preserve the full import path after KO_DOCKER_REPO.
      --push                           Push images to KO_DOCKER_REPO (default true)
  -R, --recursive                      Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (DEPRECATED)
      --sbom string                    The SBOM media type to use (none will disable SBOM synthesis and upload, also supports: spdx, cyclonedx, go.version-m). (default "spdx")
  -l, --selector string                Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
  -s, --server string                  The address and port of the Kubernetes API server (DEPRECATED)
      --tag-only                       Include tags but not digests in resolved image references. Useful when digests are not preserved when images are repopulated.
  -t, --tags strings                   Which tags to use for the produced image instead of the default 'latest' tag (may not work properly with --base-import-paths or --bare). (default [latest])
      --tarball string                 File to save images tarballs
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used (DEPRECATED)
      --token string                   Bearer token for authentication to the API server (DEPRECATED)
      --user string                    The name of the kubeconfig user to use (DEPRECATED)
      --username string                Username for basic authentication to the API server (DEPRECATED)
```

### Options inherited from parent commands

```
  -v, --verbose   Enable debug logs
```

### SEE ALSO

* [ko](ko.md)	 - Rapidly iterate with Go, Containers, and Kubernetes.

