# Configuration

## Basic Configuration

Aside from certain environment variables (see [below](#environment-variables-advanced)) like `KO_DOCKER_REPO`, you can
configure `ko`'s behavior using a `.ko.yaml` file. The location of this file can be overridden with `KO_CONFIG_PATH`.

### Overriding Base Images

By default, `ko` bases images on `cgr.dev/chainguard/static`. This is a
small image that provides the bare necessities to run your Go binary.

You can override this base image in two ways:

1. To override the base image for all images `ko` builds, add this line to your
   `.ko.yaml` file:

```yaml
defaultBaseImage: registry.example.com/base/image
```

You can also use the `KO_DEFAULTBASEIMAGE` environment variable to set the default base image, which overrides the YAML configuration:

```shell
KO_DEFAULTBASEIMAGE=registry.example.com/base/image ko build .
```

2. To override the base image for certain importpaths:

```yaml
baseImageOverrides:
  github.com/my-user/my-repo/cmd/app: registry.example.com/base/for/app
  github.com/my-user/my-repo/cmd/foo: registry.example.com/base/for/foo
```

### Overriding Go build settings

By default, `ko` builds the binary with no additional build flags other than
`-trimpath`. You can replace the default build
arguments by providing build flags and ldflags using a
[GoReleaser](https://github.com/goreleaser/goreleaser) influenced `builds`
configuration section in your `.ko.yaml`.

```yaml
builds:
- id: foo
  dir: .  # default is .
  main: ./foobar/foo
  env:
  - GOPRIVATE=git.internal.example.com,source.developers.google.com
  flags:
  - -tags
  - netgo
  ldflags:
  - -s -w
  - -extldflags "-static"
  - -X main.version={{.Env.VERSION}}
- id: bar
  dir: ./bar
  main: .  # default is .
  env:
  - GOCACHE=/workspace/.gocache
  ldflags:
  - -s
  - -w
```

If your repository contains multiple modules (multiple `go.mod` files in
different directories), use the `dir` field to specify the directory where
`ko` should run `go build`.

`ko` picks the entry from `builds` based on the import path you request. The
import path is matched against the result of joining `dir` and `main`.

The paths specified in `dir` and `main` are relative to the working directory
of the `ko` process.

The `ldflags` default value is `[]`.

### Templating support

The `ko` builds supports templating of `flags` and `ldflags`, similar to the
[GoReleaser `builds` section](https://goreleaser.com/customization/build/).

The table below lists the supported template parameters.

| Template param        | Description                                              |
|-----------------------|----------------------------------------------------------|
| `Env`                 | Map of environment variables used for the build          |
| `GoEnv`               | Map of `go env` environment variables used for the build |
| `Date`                | The UTC build date in RFC 3339 format                    |
| `Timestamp`           | The UTC build date as Unix epoc seconds                  |
| `Git.Branch`          | The current git branch                                   |
| `Git.Tag`             | The current git tag                                      |
| `Git.ShortCommit`     | The git commit short hash                                |
| `Git.FullCommit`      | The git commit full hash                                 |
| `Git.CommitDate`      | The UTC commit date in RFC 3339 format                   |
| `Git.CommitTimestamp` | The UTC commit date in Unix format                       |
| `Git.IsDirty`         | Whether or not current git state is dirty                |
| `Git.IsClean`         | Whether or not current git state is clean.               |
| `Git.TreeState`       | Either `clean` or `dirty`                                |

### Setting default platforms

By default, `ko` builds images based on the platform it runs on. If your target platform differs from your build platform you can specify the build platform:

**As a parameter**
See [Multi-Platform Images](./features/multi-platform.md).

**In .ko.yaml**
Add this to your `.ko.yaml` file:

```yaml
defaultPlatforms:
- linux/arm64
- linux/amd64
```

You can also use the `KO_DEFAULTPLATFORMS` environment variable to set the default platforms, which overrides the YAML configuration:

```shell
KO_DEFAULTPLATFORMS=linux/arm64,linux/amd64
```

### Setting build environment variables

By default, `ko` builds use the ambient environment from the system (i.e. `os.Environ()`).
These values can be overridden for your build.

```yaml
defaultEnv:
- FOO=foo
builds:
- id: foo
  dir: .
  main: ./foobar/foo
  env:
  - FOO=bar
- id: bar     # Will use defaultEnv.
  dir: ./bar
  main: .
```

For a given build, the environment variables are merged in the following order:

- System `os.Environ` (lowest precedence)
- Build variables: `build.env` if specified, otherwise `defaultEnv` (highest precedence)

### Setting build flags and ldflags

You can specify both `flags` and `ldflags` globally as well as per-build.

```yaml
defaultFlags:
- -v
defaultLdflags:
- -s
builds:
- id: foo
  dir: .
  main: ./foobar/foo
  flags:
  - -trimpath
  ldflags:
  - -w
- id: bar     # Will use defaultFlags and defaultLdflags.
  dir: ./bar
  main: .
```

The values for a `build` will be used if specified, otherwise their respective defaults will be used.
Both default and per-build values may use [template parameters](#templating-support).

### Environment Variables (advanced)

For ease of use, backward compatibility and advanced use cases, `ko` supports the following environment variables to
influence the build process.

| Variable         | Default Value                              | Description                                                                                                                                                                                                                      |
|------------------|--------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `KO_DOCKER_REPO` | (not set)                                  | Container repository where to push images built with `ko` (required)                                                                                                                                                             |
| `KO_GO_PATH`     | `go`                                       | `go` binary to use for builds, relative or absolute path, otherwise looked up via $PATH (optional)                                                                                                                               |
| `KO_CONFIG_PATH` | `./.ko.yaml`                               | Path to `ko` configuration file (optional)                                                                                                                                                                                       |
| `KOCACHE`        | (not set)                                  | This tells `ko` to store a local mapping between the `go build` inputs to the image layer that they produce, so `go build` can be skipped entirely if the layer is already present in the image registry (optional).             |

## Overwriting which directory binaries are placed in

By default `ko` places executable binaries of your application into `/ko-app`.

As it is sometimes desirable to place an executable into a specific folder inside
the container, or into a default executable path like `/bin` or `/usr/local/bin`,
this directory can be overwritten.

This custom path can be defined either
- via the `--app-dir` command line flag
- or a `defaultAppDirectory` entry in the `.ko.yaml` file.

## Naming Images

`ko` provides a few different strategies for naming the image it pushes, to
workaround certain registry limitations and user preferences:

Given `KO_DOCKER_REPO=registry.example.com/repo`, by default,
`ko build ./cmd/app` will produce an image named like
`registry.example.com/repo/app-<md5>`, which includes the MD5 hash of the full
import path, to avoid collisions.

- `--preserve-import-paths` (`-P`) will include the entire importpath:
  `registry.example.com/repo/github.com/my-user/my-repo/cmd/app`
- `--base-import-paths` (`-B`) will omit the MD5 portion:
  `registry.example.com/repo/app`
- `--bare` will only include the `KO_DOCKER_REPO`: `registry.example.com/repo`

## Local Publishing Options

`ko` is normally used to publish images to container image registries,
identified by `KO_DOCKER_REPO`.

`ko` can also load images to a local Docker daemon, if available, by setting
`KO_DOCKER_REPO=ko.local`, or by passing the `--local` (`-L`) flag.

Local images can be used as a base image for other `ko` images:

```yaml
defaultBaseImage: ko.local/example/base/image
```

`ko` can also load images into a local [KinD](https://kind.sigs.k8s.io)
cluster, if available, by setting `KO_DOCKER_REPO=kind.local`. By default this
loads into the default KinD cluster name (`kind`). To load into another KinD
cluster, set `KIND_CLUSTER_NAME=my-other-cluster`.

