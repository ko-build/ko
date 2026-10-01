# Installation

### Install from [GitHub Releases](https://github.com/ko-build/ko/releases)

```
$ VERSION=TODO # choose the latest version (without v prefix)
$ OS=Linux     # or Darwin
$ ARCH=x86_64  # or arm64, i386, s390x
```

We generate [GitHub Artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations) using [actions/attest](https://github.com/actions/attest). To verify our release, install the [GitHub CLI](https://cli.github.com) and verify as follows:

```shell
$ curl -sSfL "https://github.com/ko-build/ko/releases/download/v${VERSION}/ko_${VERSION}_${OS}_${ARCH}.tar.gz" > ko.tar.gz
$ gh attestation verify ko.tar.gz \
  --repo "ko-build/ko" \
  --signer-workflow "ko-build/ko/.github/workflows/release.yml" \
  --source-ref "refs/tags/v${VERSION}"
```

Releases up to v0.19.1 were published with [SLSA provenance](https://slsa.dev) (`multiple.intoto.jsonl`) instead of GitHub Artifact attestations. Verify those with [slsa-verifier](https://github.com/slsa-framework/slsa-verifier#installation):

```shell
$ curl -sSfL https://github.com/ko-build/ko/releases/download/v${VERSION}/multiple.intoto.jsonl > multiple.intoto.jsonl
$ slsa-verifier verify-artifact --provenance-path multiple.intoto.jsonl --source-uri github.com/ko-build/ko --source-tag "v${VERSION}" ko.tar.gz
```

```shell
$ tar xzf ko.tar.gz ko
$ chmod +x ./ko
```

### Install using [Homebrew](https://brew.sh)

```plaintext
brew install ko
```

### Install using [MacPorts](https://www.macports.org)

```plaintext
sudo port install ko
```

More info [here](https://ports.macports.org/port/ko/)

### Install on Windows using [Scoop](https://scoop.sh)

```plaintext
scoop install ko
```

### Install on [Alpine Linux](https://www.alpinelinux.org)

Installation on Alpine requires using the [`testing` repository](https://wiki.alpinelinux.org/wiki/Enable_Community_Repository#Using_testing_repositories)

```
echo https://dl-cdn.alpinelinux.org/alpine/edge/testing/ >> /etc/apk/repositories
apk update
apk add ko
```

### Build and Install from source

With Go 1.16+, build and install the latest released version:

```plaintext
go install github.com/google/ko@latest
```

### Setup on GitHub Actions

You can use the [setup-ko](https://github.com/ko-build/setup-ko) action to install ko and setup auth to [GitHub Container Registry](https://github.com/features/packages) in a GitHub Action workflow:

```plaintext
steps:
- uses: ko-build/setup-ko@v0.6
```
