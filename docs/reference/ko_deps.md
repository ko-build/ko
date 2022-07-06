## ko deps

Print Go module dependency information about the ko-built binary in the image

### Synopsis

This sub-command finds and extracts the executable binary in the image, assuming it was built by ko, and prints information about the Go module dependencies of that executable, as reported by "go version -m".

If the image was not built using ko, or if it was built without embedding dependency information, this command will fail.

```
ko deps IMAGE [flags]
```

### Examples

```

  # Fetch and extract Go dependency information from an image:
  ko deps docker.io/my-user/my-image:v3
```

### Options

```
  -h, --help          help for deps
      --sbom string   Format for SBOM output (supports: spdx, cyclonedx, go.version-m). (default "spdx")
```

### Options inherited from parent commands

```
  -v, --verbose   Enable debug logs
```

### SEE ALSO

* [ko](ko.md)	 - Rapidly iterate with Go, Containers, and Kubernetes.

