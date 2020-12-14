# ko

`ko` is a tool for building and deploying Golang applications to Kubernetes.

[![Travis Build Status](https://travis-ci.org/google/ko.svg?branch=master)](https://travis-ci.org/google/ko)
[![GitHub Actions Build Status](https://github.com/google/ko/workflows/Build/badge.svg)](https://github.com/google/ko/actions?query=workflow%3ABuild)
[![GoDoc](https://godoc.org/github.com/google/ko?status.svg)](https://godoc.org/github.com/google/ko)
[![Go Report Card](https://goreportcard.com/badge/google/ko)](https://goreportcard.com/report/google/ko)

<img src="./logo/ko.png" width="300">

## Installation

`ko` can be installed and upgraded by running:

**Note**: Golang version `1.12.0` or higher is required.

```shell
GO111MODULE=on go get github.com/google/ko/cmd/ko
```

## Authenticating

The `ko` CLI makes extensive use of the container registry as a ubiquitous and
standard object store. However, the typical model for authenticating with a
container registry is via `docker login`, and `ko` does not require users to
install `docker` locally. To facilitate logging in without `docker` we expose:

```shell
ko auth login my.registry.io -u username --password-stdin
```

## The `ko` Model

`ko` is built around a very simple extension to Go's model for expressing
dependencies using [import paths](https://golang.org/doc/code.html#ImportPaths).

In Go, dependencies are expressed via blocks like:

```go
import (
    "github.com/google/foo/pkg/hello"
    "github.com/google/bar/pkg/world"
)
```

Similarly (as you can see above), Go binaries can be referenced via import paths
like `github.com/google/ko/cmd`.

**One of the goals of `ko` is to make containers invisible infrastructure.**
Simply replace image references in your Kubernetes yaml with the import path for
your Go binary prefixed with `ko://` (e.g. `ko://github.com/google/ko/cmd/ko`),
and `ko` will handle containerizing and publishing that container image as
needed.

For example, you might use the following in a Kubernetes `Deployment` resource:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world
spec:
  selector:
    matchLabels:
      foo: bar
  replicas: 1
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
        - name: hello-world
          # This is the import path for the Go binary to build and run.
          image: ko://github.com/mattmoor/examples/http/cmd/helloworld
          ports:
            - containerPort: 8080
```

### What gets built?

`ko` will attempt to containerize and build any string within the yaml prefixed
with `ko://`.

### Results

Employing this convention enables `ko` to have effectively zero configuration
and enables very fast development iteration. For
[warm-image](https://github.com/mattmoor/warm-image), `ko` is able to build,
containerize, and redeploy a non-trivial Kubernetes controller app in seconds
(dominated by two `go build`s).

```shell
$ ko apply -f config/
2018/07/19 14:56:41 Using base gcr.io/distroless/static:nonroot for github.com/mattmoor/warm-image/cmd/sleeper
2018/07/19 14:56:42 Publishing us.gcr.io/my-project/sleeper-ebdb8b8b13d4bbe1d3592de055016d37:latest
2018/07/19 14:56:43 mounted blob: sha256:57752e7f9593cbfb7101af994b136a369ecc8174332866622db32a264f3fbefd
2018/07/19 14:56:43 mounted blob: sha256:59df9d5b488aea2753ab7774ae41a9a3e96903f87ac699f3505960e744f36f7d
2018/07/19 14:56:43 mounted blob: sha256:739b3deec2edb17c512f507894c55c2681f9724191d820cdc01f668330724ca7
2018/07/19 14:56:44 us.gcr.io/my-project/sleeper-ebdb8b8b13d4bbe1d3592de055016d37:latest: digest: sha256:6c7b96a294cad3ce613aac23c8aca5f9dd12a894354ab276c157fb5c1c2e3326 size: 592
2018/07/19 14:56:44 Published us.gcr.io/my-project/sleeper-ebdb8b8b13d4bbe1d3592de055016d37@sha256:6c7b96a294cad3ce613aac23c8aca5f9dd12a894354ab276c157fb5c1c2e3326
2018/07/19 14:56:45 Using base gcr.io/distroless/static:nonroot for github.com/mattmoor/warm-image/cmd/controller
2018/07/19 14:56:46 Publishing us.gcr.io/my-project/controller-9e91872fd7c48124dbe6ea83944b87e9:latest
2018/07/19 14:56:46 mounted blob: sha256:007782ba6738188a59bf21b4d8e974f218615ee948c6357535d07e7248b2a560
2018/07/19 14:56:46 mounted blob: sha256:57752e7f9593cbfb7101af994b136a369ecc8174332866622db32a264f3fbefd
2018/07/19 14:56:46 mounted blob: sha256:7fec050f965d7fba3de4bd19739746dce5a5125331b7845bf02185ff5d4cc374
2018/07/19 14:56:47 us.gcr.io/my-project/controller-9e91872fd7c48124dbe6ea83944b87e9:latest: digest: sha256:5a81029bb0cfd519c321aeeea2bc1b7dc6488b6c72003d3613442b4d5e4ed14d size: 593
2018/07/19 14:56:47 Published us.gcr.io/my-project/controller-9e91872fd7c48124dbe6ea83944b87e9@sha256:5a81029bb0cfd519c321aeeea2bc1b7dc6488b6c72003d3613442b4d5e4ed14d
namespace/warmimage-system configured
clusterrolebinding.rbac.authorization.k8s.io/warmimage-controller-admin configured
deployment.apps/warmimage-controller unchanged
serviceaccount/warmimage-controller unchanged
customresourcedefinition.apiextensions.k8s.io/warmimages.mattmoor.io configured
```

## Usage

`ko` has four commands, most of which build and publish images as part of their
execution. By default, `ko` publishes images to a Docker Registry specified via
`KO_DOCKER_REPO`.

**Note**: You'll need to be authenticated with your `KO_DOCKER_REPO` before
pushing images. Run `gcloud auth configure-docker` if you are using Google
Container Registry or `docker login` if you are using Docker Hub.

However, these same commands can be directed to operate locally as well via the
`--local` or `-L` command (or setting `KO_DOCKER_REPO=ko.local`). See the
[`minikube` section](./README.md#with-minikube) for more detail.

`ko` can also be used with `kind` directly by setting
`KO_DOCKER_REPO=kind.local`. See [the relevant section](./README.md#with-kind)
for more detail.

### `ko publish`

`ko publish` simply builds and publishes images for each import path passed as
an argument. It prints the images' published digests after each image is
published.

```shell
$ ko publish github.com/mattmoor/warm-image/cmd/sleeper
2018/07/19 14:57:34 Using base gcr.io/distroless/static:nonroot for github.com/mattmoor/warm-image/cmd/sleeper
2018/07/19 14:57:35 Publishing us.gcr.io/my-project/sleeper-ebdb8b8b13d4bbe1d3592de055016d37:latest
2018/07/19 14:57:35 mounted blob: sha256:739b3deec2edb17c512f507894c55c2681f9724191d820cdc01f668330724ca7
2018/07/19 14:57:35 mounted blob: sha256:57752e7f9593cbfb7101af994b136a369ecc8174332866622db32a264f3fbefd
2018/07/19 14:57:35 mounted blob: sha256:59df9d5b488aea2753ab7774ae41a9a3e96903f87ac699f3505960e744f36f7d
2018/07/19 14:57:36 us.gcr.io/my-project/sleeper-ebdb8b8b13d4bbe1d3592de055016d37:latest: digest: sha256:6c7b96a294cad3ce613aac23c8aca5f9dd12a894354ab276c157fb5c1c2e3326 size: 592
2018/07/19 14:57:36 Published us.gcr.io/my-project/sleeper-ebdb8b8b13d4bbe1d3592de055016d37@sha256:6c7b96a294cad3ce613aac23c8aca5f9dd12a894354ab276c157fb5c1c2e3326
```

`ko publish` also supports relative import paths, when in the context of a repo
on `GOPATH`.

```shell
$ ko publish ./cmd/sleeper
2018/07/19 14:58:16 Using base gcr.io/distroless/static:nonroot for github.com/mattmoor/warm-image/cmd/sleeper
2018/07/19 14:58:16 Publishing us.gcr.io/my-project/sleeper-ebdb8b8b13d4bbe1d3592de055016d37:latest
2018/07/19 14:58:17 mounted blob: sha256:59df9d5b488aea2753ab7774ae41a9a3e96903f87ac699f3505960e744f36f7d
2018/07/19 14:58:17 mounted blob: sha256:739b3deec2edb17c512f507894c55c2681f9724191d820cdc01f668330724ca7
2018/07/19 14:58:17 mounted blob: sha256:57752e7f9593cbfb7101af994b136a369ecc8174332866622db32a264f3fbefd
2018/07/19 14:58:18 us.gcr.io/my-project/sleeper-ebdb8b8b13d4bbe1d3592de055016d37:latest: digest: sha256:6c7b96a294cad3ce613aac23c8aca5f9dd12a894354ab276c157fb5c1c2e3326 size: 592
2018/07/19 14:58:18 Published us.gcr.io/my-project/sleeper-ebdb8b8b13d4bbe1d3592de055016d37@sha256:6c7b96a294cad3ce613aac23c8aca5f9dd12a894354ab276c157fb5c1c2e3326
```

### `ko resolve`

`ko resolve` takes Kubernetes yaml files in the style of `kubectl apply` and
(based on the [model above](#the-ko-model)) determines the set of Go import
paths to build, containerize, and publish.

The output of `ko resolve` is the concatenated yaml with import paths replaced
with published image digests. Following the example above, this would be:

```shell
# Command
export PROJECT_ID=$(gcloud config get-value core/project)
export KO_DOCKER_REPO="gcr.io/${PROJECT_ID}"
ko resolve -f deployment.yaml

# Output
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: hello-world
        # This is the digest of the published image containing the go binary.
        image: gcr.io/your-project/helloworld-badf00d@sha256:deadbeef
        ports:
        - containerPort: 8080
```

Some Docker Registries (e.g. gcr.io) support multi-level repository names. For
these registries, it is often useful for discoverability and provenance to
preserve the full import path, for this we expose `--preserve-import-paths`, or
`-P` for short.

```shell
# Command
export PROJECT_ID=$(gcloud config get-value core/project)
export KO_DOCKER_REPO="gcr.io/${PROJECT_ID}"
ko resolve -P -f deployment.yaml

# Output
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: hello-world
        # This is the digest of the published image containing the go binary
        # at the embedded import path.
        image: gcr.io/your-project/github.com/mattmoor/examples/http/cmd/helloworld@sha256:deadbeef
        ports:
        - containerPort: 8080
```

It is notable that this is not the default (anymore) because certain popular
registries (including Docker Hub) do not support multi-level repository names.

`ko resolve`, `ko apply`, and `ko create` accept an optional `--selector` or
`-l` flag, similar to `kubectl`, which can be used to filter the resources from
the input Kubernetes YAMLs by their `metadata.labels`.

In the case of `ko resolve`, `--selector` will render only the resources that
are selected by the provided selector.

See
[the documentation on Kubernetes selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
for more information on using label selectors.

### `ko apply`

`ko apply` is intended to parallel `kubectl apply`, but acts on the same
resolved output as `ko resolve` emits. It is expected that `ko apply` will act
as the vehicle for rapid iteration during development. As changes are made to a
particular application, you can run: `ko apply -f unit.yaml` to rapidly rebuild,
repush, and redeploy their changes.

`ko apply` will invoke `kubectl apply` under the hood, and therefore apply to
whatever `kubectl` context is active.

### `ko apply --watch` (EXPERIMENTAL)

The `--watch` flag (`-W` for short) does an initial `apply` as above, but as it
does, it builds up a dependency graph of your program and starts to continuously
monitor the filesystem for changes. When a file changes, it re-applies any yamls
that are affected.

For example, if I edit `github.com/foo/bar/pkg/baz/blah.go`, the tool sees that
the `github.com/foo/bar/pkg/baz` package has changed, and perhaps both
`github.com/foo/bar/cmd/one` and `github.com/foo/bar/cmd/two` consume that
library and were referenced by `config/one-deploy.yaml` and
`config/two-deploy.yaml`. The edit would effectively result in a re-application
of:

```
ko apply -f config/one-deploy.yaml -f config/two-deploy.yaml
```

This flag is still experimental, and feedback is very welcome.

### `ko delete`

`ko delete` simply passes through to `kubectl delete`. It is exposed purely out
of convenience for cleaning up resources created through `ko apply`.

### `ko version`

`ko version` prints version of ko. For not released binaries it will print hash
of latest commit in current git tree.

## With `minikube`

You can use `ko` with `minikube` via a Docker Registry, but this involves
publishing images only to pull them back down to your machine again. To avoid
this, `ko` exposes `--local` or `-L` options to instead publish the images to
the local machine's Docker daemon.

This would look something like:

```shell
# Use the minikube docker daemon.
eval $(minikube docker-env)

# Make sure minikube is the current kubectl context.
kubectl config use-context minikube

# Deploy to minikube w/o registry.
ko apply -L -f config/

# This is the same as above.
KO_DOCKER_REPO=ko.local ko apply -f config/
```

A caveat of this approach is that it will not work if your container is
configured with `imagePullPolicy: Always` because despite having the image
locally, a pull is performed to ensure we have the latest version, it still
exists, and that access hasn't been revoked. A workaround for this is to use
`imagePullPolicy: IfNotPresent`, which should work well with `ko` in all
contexts.

Images will appear in the Docker daemon as
`ko.local/import.path.com/foo/cmd/bar`. With `--local` import paths are always
preserved (see `--preserve-import-paths`).

## With `kind`

Likewise, you can use `ko` with [kind](https://github.com/kubernetes-sigs/kind)
to aid in rapid local iteration both locally and in small CI environments. To
instruct `ko` to publish images into your `kind` cluster, the `KO_DOCKER_REPO`
variable must be set to `kind.local`.

This would look something like:

```shell
# Create a kind cluster
kind create cluster

# Deploy to kind w/o registry.
KO_DOCKER_REPO=kind.local ko apply -f config/
```

If you want to create a `kind` cluster with a non default name, you can set the
`KIND_CLUSTER_NAME` variable to the respective name (which is also supported by
[`kind` itself](https://github.com/kubernetes-sigs/kind/releases/tag/v0.8.0)).

Like with `minikube` above, a caveat of this approach is that it will not work
if your container is configured with `imagePullPolicy: Always` because despite
having the image locally, a pull is performed to ensure we have the latest
version, it still exists, and that access hasn't been revoked. A workaround for
this is to use `imagePullPolicy: IfNotPresent`, which should work well with `ko`
in all contexts.

Note that images will not appear in the Docker daemon running `kind` as the
cluster itself is running in a container that is running `containerd` inside.
The images are loaded into the respective `containerd` daemon.

## Configuration via `.ko.yaml`

While `ko` aims to have zero configuration, there are certain scenarios where
you will want to override `ko`'s default behavior. This is done via `.ko.yaml`.

`.ko.yaml` is put into the directory from which `ko` will be invoked. One can
override the directory with the `KO_CONFIG_PATH` environment variable.

If neither is present, then `ko` will rely on its default behaviors.

### Overriding the default base image

By default, `ko` makes use of `gcr.io/distroless/static:nonroot` as the base
image for containers. There are a wide array of scenarios in which overriding
this makes sense, for example:

1. Pinning to a particular digest of this image for repeatable builds,
1. Replacing this streamlined base image with another with better debugging
   tools (e.g. a shell, like `docker.io/library/ubuntu`).

The default base image `ko` uses can be changed by simply adding the following
line to `.ko.yaml`:

```yaml
defaultBaseImage: gcr.io/another-project/another-image@sha256:deadbeef
```

### Overriding the base for particular imports

Some of your binaries may have requirements that are a more unique, and you may
want to direct `ko` to use a particular base image for just those binaries.

The base image `ko` uses can be changed by adding the following to `.ko.yaml`:

```yaml
baseImageOverrides:
  github.com/my-org/my-repo/path/to/binary: docker.io/another/base:latest
```

### Why isn't `KO_DOCKER_REPO` part of `.ko.yaml`?

Once introduced to `.ko.yaml`, you may find yourself wondering: Why does it not
hold the value of `$KO_DOCKER_REPO`?

The answer is that `.ko.yaml` is expected to sit in the root of a repository,
and get checked in and versioned alongside your source code. This also means
that the configured values will be shared across developers on a project, which
for `KO_DOCKER_REPO` is actually undesirable because each developer is (likely)
using their own docker repository and cluster.

## Including static assets

A question that often comes up after using `ko` for a while is: "How do I
include static assets in images produced with `ko`?".

For this, `ko` builds around an idiom similar to `go test` and `testdata/`. `ko`
will include all of the data under `<import path>/kodata/...` in the images it
produces.

These files are placed under `/var/run/ko/...`, but the appropriate mechanism
for referencing them should be through the `KO_DATA_PATH` environment variable.
The intent of this is to enable users to test things outside of `ko` as follows:

```shell
KO_DATA_PATH=$PWD/cmd/ko/test/kodata go run ./cmd/ko/test/*.go
2018/07/19 23:35:20 Hello there
```

This produces identical output to being run within the container locally:

```shell
ko publish -L ./cmd/test
2018/07/19 23:36:11 Using base gcr.io/distroless/static:nonroot for github.com/google/ko/cmd/test
2018/07/19 23:36:12 Loading ko.local/github.com/google/ko/cmd/test:703c205bf2f405af520b40536b87aafadcf181562b8faa6690fd2992084c8577
2018/07/19 23:36:13 Loaded ko.local/github.com/google/ko/cmd/test:703c205bf2f405af520b40536b87aafadcf181562b8faa6690fd2992084c8577

docker run -ti --rm ko.local/github.com/google/ko/cmd/test:703c205bf2f405af520b40536b87aafadcf181562b8faa6690fd2992084c8577
2018/07/19 23:36:25 Hello there
```

... or on cluster:

```shell
ko apply -f cmd/ko/test/test.yaml
2018/07/19 23:38:24 Using base gcr.io/distroless/static:nonroot for github.com/google/ko/cmd/test
2018/07/19 23:38:25 Publishing us.gcr.io/my-project/test-294a7bdc57d85dc6ddeef5ba38a59fe9:latest
2018/07/19 23:38:26 mounted blob: sha256:988abcba36b5948da8baa1e3616b94c0b56da814b8f6ff3ae3ac316e375e093a
2018/07/19 23:38:26 mounted blob: sha256:57752e7f9593cbfb7101af994b136a369ecc8174332866622db32a264f3fbefd
2018/07/19 23:38:26 mounted blob: sha256:f24d43c24e22298ed99ea125af6c1b828ae07716968f78cb6d09d4291a13f2d3
2018/07/19 23:38:26 mounted blob: sha256:7a7bafbc2ae1bf844c47b33025dd459913a3fece0a94b1f3ced860675be2b79c
2018/07/19 23:38:27 us.gcr.io/my-project/test-294a7bdc57d85dc6ddeef5ba38a59fe9:latest: digest: sha256:703c205bf2f405af520b40536b87aafadcf181562b8faa6690fd2992084c8577 size: 751
2018/07/19 23:38:27 Published us.gcr.io/my-project/test-294a7bdc57d85dc6ddeef5ba38a59fe9@sha256:703c205bf2f405af520b40536b87aafadcf181562b8faa6690fd2992084c8577
pod/kodata created

kubectl logs kodata
2018/07/19 23:38:29 Hello there
```

## Multi-Platform Images

If `ko` is invoked with `--platform=all`, for any image that it builds that is
based on a multi-architecture image (e.g., the default
`gcr.io/distroless/static:nonroot`, `busybox`, `alpine`, etc.), `ko` will
attempt to build the Go binary using Go's cross-compilation support and produce
a multi-architecture
[image index](https://github.com/opencontainers/image-spec/blob/master/image-index.md)
(aka "manifest list"), with support for each OS and architecture pair supported
by the base image.

If `ko` is invoked with `--platform=<some-OS>/<some-platform>` (e.g.,
`--platform=linux/amd64` or `--platform=linux/arm64`), then it will attempt to
build an image for that OS and architecture only, assuming the base image
supports it.

When `--platform` is not provided, `ko` builds an image with the OS and
architecture based on the build environment's `GOOS` and `GOARCH`.

## Enable Autocompletion

To generate an bash completion script, you can run:

```
ko completion
```

To use the completion script, you can copy the script in your bash_completion
directory (e.g. /usr/local/etc/bash_completion.d/):

```
ko completion > /usr/local/etc/bash_completion.d/ko
```

or source it in your shell by running:

```
source <(ko completion)
```

## Relevance to Release Management

`ko` is also useful for helping manage releases. For example, if your project
periodically releases a set of images and configuration to launch those images
on a Kubernetes cluster, release binaries may be published and the configuration
generated via:

```shell
export PROJECT_ID=<YOUR RELEASE PROJECT>
export KO_DOCKER_REPO="gcr.io/${PROJECT_ID}"
ko resolve -f config/ > release.yaml
```

> Note that in this context it is recommended that you also provide `-P`, if
> supported by your Docker registry. This improves users' ability to tie release
> binaries back to their source.

This will publish all of the binary components as container images to
`gcr.io/my-releases/...` and create a `release.yaml` file containing all of the
configuration for your application with inlined image references.

This resulting configuration may then be installed onto Kubernetes clusters via:

```shell
kubectl apply -f release.yaml
```

### Why are my images all created in 1970?

In order to support [reproducible builds](https://reproducible-builds.org), `ko`
doesn't embed timestamps in the images it produces by default; however, `ko`
does respect the
[`SOURCE_DATE_EPOCH`](https://reproducible-builds.org/docs/source-date-epoch/)
environment variable.

For example, you can set this to the current timestamp by executing:

    export SOURCE_DATE_EPOCH=$(date +%s)

or to the latest git commit's timestamp with:

    export SOURCE_DATE_EPOCH=$(git log -1 --format='%ct')

## Acknowledgements

This work is based heavily on learnings from having built the
[Docker](https://github.com/bazelbuild/rules_docker) and
[Kubernetes](https://github.com/bazelbuild/rules_k8s) support for
[Bazel](https://bazel.build). That work was presented
[here](https://www.youtube.com/watch?v=RS1aiQqgUTA).
