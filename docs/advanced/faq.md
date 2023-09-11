# Frequently Asked Questions

## How can I set `ldflags`?

[Using -ldflags](https://blog.cloudflare.com/setting-go-variables-at-compile-time/) is a common way to embed version info in go binaries (In fact, we do this for `ko`!).
Unfortunately, because `ko` wraps `go build`, it's not possible to use this flag directly; however, you can use the `GOFLAGS` environment variable instead:

```sh
GOFLAGS="-ldflags=-X=main.version=1.2.3" ko build .
```

Currently, there is a limitation that does not allow to set multiple arguments in `ldflags` using `GOFLAGS`.
Using `-ldflags` multiple times also does not work.
In this use case, it works best to use the [`builds` section](./../configuration) in the `.ko.yaml` file.

## Why are my images all created in 1970?

In order to support [reproducible builds](https://reproducible-builds.org), `ko` doesn't embed timestamps in the images it produces by default.

However, `ko` does respect the [`SOURCE_DATE_EPOCH`](https://reproducible-builds.org/docs/source-date-epoch/) environment variable, which will set the container image's timestamp accordingly.

Similarly, the `KO_DATA_DATE_EPOCH` environment variable can be used to set the _modtime_ timestamp of the files in `KO_DATA_PATH`.

For example, you can set the container image's timestamp to the current timestamp by executing:

```sh
export SOURCE_DATE_EPOCH=$(date +%s)
```

or set the timestamp of the files in `KO_DATA_PATH` to the latest git commit's timestamp with:

```sh
export KO_DATA_DATE_EPOCH=$(git log -1 --format='%ct')
```

## Can I build Windows containers?

Yes, but support for Windows containers is new, experimental, and tenuous. Be prepared to file bugs. 🐛

The default base image does not provide a Windows image.
You can try out building a Windows container image by [setting the base image](./../configuration) to a Windows base image and building with `--platform=windows/amd64` or `--platform=all`:

For example, to build a Windows container image, update your `.ko.yaml` to set the base image:

```plaintext
defaultBaseImage: mcr.microsoft.com/windows/nanoserver:ltsc2022
```

And build for `windows/amd64`.

```sh
ko build ./ --platform=windows/amd64
```

### Known issues 🐛

- Symlinks in `kodata` are ignored when building Windows images; only regular files and directories will be included in the Windows image.

## Does `ko` support autocompletion?

Yes! `ko completion` generates a Bash/Zsh/Fish/PowerShell completion script.
You can get how to load it from help document.

```sh
ko completion [bash|zsh|fish|powershell] --help
```

Or, you can source it directly:

```bash
source <(ko completion)
```

## Does `ko` work with [Kustomize](https://kustomize.io/)?

Yes! `ko resolve -f -` will read and process input from stdin, so you can have `ko` easily process the output of the `kustomize` command.

```sh
kustomize build config | ko resolve -f -
```

## Does `ko` integrate with other build and development tools?

Oh, you betcha. Here's a partial list:

- `ko` support in [Skaffold](https://skaffold.dev/docs/pipeline-stages/builders/ko/)
- `ko` support for [goreleaser](https://goreleaser.com/customization/ko/)
- `ko` task in the [Tekton catalog](https://github.com/tektoncd/catalog/tree/main/task/ko/)
- `ko` support in [Carvel's `kbld`](https://carvel.dev/kbld/docs/latest/config/#ko)
- `ko` extension for [Tilt](https://github.com/tilt-dev/tilt-extensions/tree/master/ko)

## Does `ko` work with [OpenShift Internal Registry](https://access.redhat.com/documentation/en-us/openshift_container_platform/4.11/html/registry/registry-overview#registry-integrated-openshift-registry_registry-overview)?

Yes! Follow these steps:

1. [Connect to your OpenShift installation](https://docs.openshift.com/container-platform/latest/cli_reference/openshift_cli/getting-started-cli.html#cli-logging-in_cli-developer-commands)
1. [Expose the OpenShift Internal Registry](https://docs.openshift.com/container-platform/latest/registry/securing-exposing-registry.html) so you can push to it:
1. Export your token to `$HOME/.docker/config.json`:

```sh
oc registry login --to=$HOME/.docker/config.json
```

1. Create a namespace where you will push your images, i.e: `ko-images`
1. Execute this command to set `KO_DOCKER_REPO` to publish images to the internal registry.

```sh
export KO_DOCKER_REPO=$(oc registry info --public)/ko-images
```

