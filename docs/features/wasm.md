# Wasm

`ko` has early **experimental** support for building wasm containers.

This currently requrires:

- `ko` built from head (`go install github.com/google/ko@main`)
- [Go 1.21](https://go.dev/dl/) or later
- [up-to-date Docker Desktop with experimental features enabled](https://docs.docker.com/desktop/wasm/)

With those requirements met, you can build a wasm container with:

```sh
ko build ./cmd/app --platform=wasip1/wasm
```

You can then run this image with:

```sh
docker run \
  --runtime=io.containerd.wasmedge.v1 \
  --platform=wasip1/wasm \
  ${IMAGE}
```

### Known Limitations

- SBOMs are not generated for containers built this way
- `--platform` must only specify `wasip1/wasm`, and no other platforms
- Test coverage is limited. [Anecdotally](https://github.com/ko-build/ko/pull/1095), the `wasmedge` and `wasmtime` runtimes should work, but `slight` and `spin` do not currently.
