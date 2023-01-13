# GitHub Action

To make `ko` even easier to use with [GitHub Actions](https://github.com/features/actions), use [`setup-ko`](https://github.com/marketplace/actions/setup-ko).

This action installs `ko`, logs in using `${{ github.token }}`, and configures `ko` to push to [GHCR](https://github.com/features/packages) by default.

## Example

```yaml
name: Publish

on:
  push:
    branches: ['main']

jobs:
  publish:
    name: Publish
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v3
      - uses: actions/checkout@v3
      - uses: ko-build/setup-ko@v0.6
      - run: ko build
```

_That's it!_ This workflow will build and publish your code to [GitHub Container Registry](https://ghcr.io).

By default, the action sets `KO_DOCKER_REPO=ghcr.io/[owner]/[repo]` for all subsequent steps, and uses the `${{ github.token }}` to authorize pushes to GHCR.

See [documentation](./configuration.md) to learn more about configuring `ko`.

The action works on Linux and macOS [runners](https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners).

## Select `ko` version to install

By default, `setup-ko` installs the [latest release of `ko`](../releases).

You can specify a specific release version:

```yaml
- uses: ko-build/setup-ko@v0.6
  with:
    version: v0.11.2
```

Or install `ko` from HEAD using `go install`:

```yaml
- uses: ko-build/setup-ko@v0.6
  with:
    version: tip
```

## Pushing to other locations

By default, `setup-ko` [configures `ko`](./configuration.md) to push images to [GitHub Container Registry](https://ghcr.io), but you can configure it to push to other registries as well.

If `KO_DOCKER_REPO` is already set when `setup-ko` runs, it will skip logging in to ghcr.io and will propagate `KO_DOCKER_REPO` for subsequent steps.

You will also need to provide credentials to authorize the push. You can use [encrypted secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets) to store the authorization token, and pass it to `ko login` before pushing:

```yaml
steps:
...
- uses: ko-build/setup-ko@v0.6
  env:
    KO_DOCKER_REPO: my.registry/my-repo
- env:
  run: |
    echo "${{ secrets.auth_token }}" | ko login https://my.registry --username my-username --password-stdin
    ko build
```

## Release Integration

`ko` can produce YAML files containing references to built images, using [`ko resolve`](./features/k8s.md).

With this action, you can use `ko resolve` to produce output YAML that you then attach to a GitHub Release using the [GitHub CLI](https://cli.github.com).

For example:

```yaml
name: Publish Release YAML

on:
  release:
    types: ['created']

jobs:
  publish-release-yaml:
    name: Publish Release YAML
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v2
        with:
          go-version: 1.15
      - uses: actions/checkout@v2
      - uses: ko-build/setup-ko@v0.6

      - name: Generate and upload release.yaml
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          tag=$(echo ${{ github.ref }} | cut -c11-)  # get tag name without tags/refs/ prefix.
          ko resolve -t ${tag} -f config/ > release.yaml
          gh release upload ${tag} release.yaml
```
