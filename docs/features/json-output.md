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
      "platforms": ["linux/amd64", "linux/arm64"]
    }
  ]
}
```

- `tags` lists `repo:tag` (no digest) for every `--tags` value (default `latest`) of every image, sorted. Images sharing a name (`--bare`) repeat. The digest is in `images[].ref` and `images[].digest`.
- `labels` and `annotations` come from `--image-label` and `--image-annotation`.
- `images` is sorted by import path.

`-O text` is the default and keeps the existing output. `--output` is only available on `ko build` (and its alias `ko publish`).
