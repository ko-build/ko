# Installation

### Install from [GitHub Releases](https://github.com/ko-build/ko/releases)

```
$ VERSION=TODO # choose the latest version (without v prefix)
$ OS=Linux     # or Darwin
$ ARCH=x86_64  # or arm64, i386, s390x
```

We generate [SLSA3 provenance](slsa.dev) using the OpenSSF's [slsa-framework/slsa-github-generator](https://github.com/slsa-framework/slsa-github-generator). To verify our release, install the verification tool from [slsa-framework/slsa-verifier#installation](https://github.com/slsa-framework/slsa-verifier#installation) and verify as follows:


```shell
$ curl -sSfL "https://github.com/ko-build/ko/releases/download/v${VERSION}/ko_${VERSION}_${OS}_${ARCH}.tar.gz" > ko.tar.gz
$ curl -sSfL https://github.com/ko-build/ko/releases/download/v${VERSION}/attestation.intoto.jsonl > provenance.intoto.jsonl
$ slsa-verifier -artifact-path ko.tar.gz -provenance provenance.intoto.jsonl -source github.com/google/ko -tag "v${VERSION}"
  PASSED: Verified SLSA provenance
```

```shell
$ tar xzf ko.tar.gz ko
$ chmod +x ./ko
```

### Install using [Homebrew](https://brew.sh)

```plaintext
brew install ko
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

You can use the [setup-ko](https://github.com/imjasonh/setup-ko) action to install ko and setup auth to [GitHub Container Registry](https://github.com/features/packages) in a GitHub Action workflow:

```plaintext
steps:
- uses: imjasonh/setup-ko@v0.6
```
