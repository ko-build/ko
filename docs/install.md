# Installation

### Install from [GitHub Releases](https://github.com/google/ko/releases)

```plaintext
VERSION=TODO # choose the latest version
OS=Linux     # or Darwin, Windows
ARCH=x86_64  # or arm64, i386, s390x
curl -L https://github.com/google/ko/releases/download/v${VERSION}/ko_${VERSION}_${OS}_${ARCH}.tar.gz | tar xzf - ko
chmod +x ./ko
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
- uses: imjasonh/setup-ko@v0.4
```

