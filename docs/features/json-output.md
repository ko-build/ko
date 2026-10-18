# Machine-readable Output

By default `ko build` prints the published image references to stdout, one per line, and progress logs to stderr.

For scripts, CI, and agents, pass `-O json` (or `--output json`) to write a single JSON document to stdout instead. Logs still go to stderr, so stdout can be piped straight into `jq`.

```
ko build -O json --image-label org.opencontainers.image.title=app ./cmd/app | jq -r '.images[0].digest'
```

The `tags`, `labels` and `annotations` keys have the same shape as the [`json` output of docker/metadata-action](https://github.com/docker/metadata-action#outputs). The other keys are `ko` extras:

```json
{
  "version": "v0.19.0",
  "tags": ["registry.example.com/app:latest"],
  "labels": {"org.opencontainers.image.title": "app"},
  "annotations": ["key=value"],
  "images": [
    {
      "importpath": "example.com/app/cmd/app",
      "ref": "registry.example.com/app:latest@sha256:...",
      "digest": "sha256:...",
      "platforms": ["linux/amd64", "linux/arm64"],
      "sbom": {
        "index": {"format": "spdx", "mediaType": "text/spdx+json", "digest": "sha256:...", "size": 5418,
                  "ref": "registry.example.com/app:sha256-<index digest>.sbom", "path": "/abs/sboms/app-index.spdx.json"},
        "platforms": [
          {"platform": "linux/amd64", "format": "spdx", "mediaType": "text/spdx+json", "digest": "sha256:...", "size": 178078,
           "ref": "registry.example.com/app:sha256-<amd64 digest>.sbom", "path": "/abs/sboms/app-linux-amd64.spdx.json"},
          {"platform": "linux/arm64", "format": "spdx", "mediaType": "text/spdx+json", "digest": "sha256:...", "size": 178078,
           "ref": "registry.example.com/app:sha256-<arm64 digest>.sbom", "path": "/abs/sboms/app-linux-arm64.spdx.json"}
        ]
      }
    }
  ]
}
```

- `tags` lists `repo:tag` (no digest) for every `--tags` value (default `latest`) of every image, sorted. Images sharing a name (`--bare`) repeat. The digest is in `images[].ref` and `images[].digest`.
- `labels` and `annotations` come from `--image-label` and `--image-annotation`.
- `images` is sorted by import path.
- `images[].sbom` references the [SBOMs](sboms.md) ko generated, without embedding them:
  - `platforms` has the SBOM of each platform image, which lists what its binary was built from.
  - `index` is the SBOM of a multi-platform image index, which relates its platform images. It is `null` when no index was published: for a single platform, and with `--local` or kind.
  - `digest` is the sha256 of the SBOM document, and `size` its size in bytes.
  - `ref` is where the SBOM was pushed, only present when images are pushed to a registry (not with `--push=false`, `--local` or kind). To fetch the document, get the blob `digest` from the repository of `ref`, e.g. `oras blob fetch --output - registry.example.com/app@sha256:...`.
  - `path` is the SBOM file written by `--sbom-dir`, only present when that file holds this very document: import paths sharing a base name overwrite each other's file.
  - It is `null` when the image has no SBOM, e.g. with `--sbom=none`: iterate with `jq '.images[].sbom.platforms[]?'`.

`-O text` is the default and keeps the existing output. `--output` is only available on `ko build` (and its alias `ko publish`).
