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
      "sbom": [
        {"format": "spdx", "mediaType": "text/spdx+json", "digest": "sha256:...", "size": 5418, "ref": "registry.example.com/app:sha256-<index digest>.sbom"},
        {"platform": "linux/amd64", "format": "spdx", "mediaType": "text/spdx+json", "digest": "sha256:...", "size": 178078, "ref": "registry.example.com/app:sha256-<amd64 digest>.sbom"},
        {"platform": "linux/arm64", "format": "spdx", "mediaType": "text/spdx+json", "digest": "sha256:...", "size": 178078, "ref": "registry.example.com/app:sha256-<arm64 digest>.sbom"}
      ]
    }
  ]
}
```

- `tags` lists `repo:tag` (no digest) for every `--tags` value (default `latest`) of every image, sorted. Images sharing a name (`--bare`) repeat. The digest is in `images[].ref` and `images[].digest`.
- `labels` and `annotations` come from `--image-label` and `--image-annotation`.
- `images` is sorted by import path.
- `images[].sbom` references the [SBOMs](sboms.md) ko generated, without embedding them:
  - The first entry without a `platform` is the SBOM of the published image or index; for a multi-platform index there is one more entry per platform image.
  - `digest` is the digest of the SBOM artifact, and `size` is the size of the SBOM document in bytes.
  - `ref` is where the SBOM was pushed in the registry, and is only present when images are pushed to a registry. It is omitted with `--push=false`, `--local` and kind.
  - It is `[]` with `--sbom=none`.

`-O text` is the default and keeps the existing output. `--output` is only available on `ko build` (and its alias `ko publish`).
