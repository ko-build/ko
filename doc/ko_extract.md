## ko extract

Extract ko-built image references from YAML configs

### Synopsis

This sub-command extracts image references detected to have been built by ko

```
ko extract -f FILENAME... [flags]
```

### Examples

```

# Extract and sign ko-built images:
ko extract -f release.yaml | xargs cosign sign ...

# Extract and rebase ko-built images:
ko extract -f release.yaml | xargs crane rebase ...

```

### Options

```
  -f, --filename strings   Filename, directory, or URL to files to use to create the resource
  -h, --help               help for extract
  -R, --recursive          Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.
  -W, --watch              Continuously monitor the transitive dependencies of the passed yaml files, and redeploy whenever anything changes. (DEPRECATED)
```

### SEE ALSO

* [ko](ko.md)	 - Rapidly iterate with Go, Containers, and Kubernetes.

