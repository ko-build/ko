# Installation

### Install from [GitHub Releases](https://github.com/ko-build/ko/releases)

```
$ VERSION=TODO # choose the latest version (without v prefix)
$ OS=Linux     # or Darwin
$ ARCH=x86_64  # or arm64, i386, s390x
```

We generate [SLSA3 provenance](https://slsa.dev) using the OpenSSF's [slsa-framework/slsa-github-generator](https://github.com/slsa-framework/slsa-github-generator). To verify our release, install the verification tool from [slsa-framework/slsa-verifier#installation](https://github.com/slsa-framework/slsa-verifier#installation) and, depending on which version of `ko` you're installing, verify as follows:

#### v0.13.0+

In the `v0.13.0` release, the `attestation.intoto.jsonl` file was renamed to `multiple.intoto.jsonl` so if you're trying to install `v0.13.0` or later, you'll need to use the `multiple.intoto.jsonl` file:


```shell
$ curl -sSfL "https://github.com/ko-build/ko/releases/download/v${VERSION}/ko_${VERSION}_${OS}_${ARCH}.tar.gz" > ko.tar.gz
$ curl -sSfL https://github.com/ko-build/ko/releases/download/v${VERSION}/multiple.intoto.jsonl > provenance.intoto.jsonl
$ slsa-verifier -artifact-path ko.tar.gz -provenance provenance.intoto.jsonl -source github.com/google/ko -tag "v${VERSION}"
  PASSED: Verified SLSA provenance
```

### v0.12.0 and earlier

For any version of `ko` before `v0.13.0`, you'll need to use the `attestation.intoto.jsonl` file:

```shell
$ curl -sSfL "https://github.com/ko-build/ko/releases/download/v${VERSION}/ko_${VERSION}_${OS}_${ARCH}.tar.gz" > ko.tar.gz
$ curl -sSfL https://github.com/ko-build/ko/releases/download/v${VERSION}/multiple.intoto.jsonl > multiple.intoto.jsonl
$ slsa-verifier verify-artifact --provenance-path multiple.intoto.jsonl --source-uri github.com/ko-build/ko --source-tag "v${VERSION}" ko.tar.gz
Verified signature against tlog entry index 24413745 at URL: https://rekor.sigstore.dev/api/v1/log/entries/24296fb24b8ad77ab97a5263b5fa8f35789618348a39358b1f9470b0c31045effbbe5e23e77a5836
Verified build using builder "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v1.7.0" at commit 200db7243f02b5c0303e21d8ab8e3b4ad3a229d0
Verifying artifact /Users/batuhanapaydin/workspace/ko/ko.tar.gz: PASSED

PASSED: Verified SLSA provenance
```

And finally, after verifying the release, you can unpack the binary and make it executable:

```shell
$ tar xzf ko.tar.gz ko
$ chmod +x ./ko
```

### Install using [Homebrew](https://brew.sh)

```plaintext
brew install ko
```

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

You can use the [setup-ko](https://github.com/imjasonh/setup-ko) action to install ko and setup auth to [GitHub Container Registry](https://github.com/features/packages) in a GitHub Action workflow:

```plaintext
steps:
- uses: imjasonh/setup-ko@v0.6
```
