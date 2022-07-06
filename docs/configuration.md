# Configuration

## Basic Configuration

Aside from `KO_DOCKER_REPO`, you can configure `ko`'s behavior using a
`.ko.yaml` file. The location of this file can be overridden with
`KO_CONFIG_PATH`.

### Overriding Base Images

By default, `ko` bases images on `gcr.io/distroless/static:nonroot`. This is a
small image that provides the bare necessities to run your Go binary.

You can override this base image in two ways:

1. To override the base image for all images `ko` builds, add this line to your
   `.ko.yaml` file:

```yaml
defaultBaseImage: registry.example.com/base/image
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

> 💡 **Note:** Even though the configuration section is similar to the
[GoReleaser `builds` section](https://goreleaser.com/customization/build/),
only the `env`, `flags` and `ldflags` fields are currently supported. Also, the
templating support is currently limited to using environment variables only.

## Naming Images

`ko` provides a few different strategies for naming the image it pushes, to
workaround certain registry limitations and user preferences:

Given `KO_DOCKER_REPO=registry.example.com/repo`, by default,
`ko build ./cmd/app` will produce an image named like
`registry.example.com/repo/app-<md5>`, which includes the MD5 hash of the full
import path, to avoid collisions.

- `--preserve-import-path` (`-P`) will include the entire importpath:
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

