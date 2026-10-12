# Debugging

Sometimes it's challenging to track down the cause of unexpected behavior in an app. Because `ko` makes it simple to make tweaks to your app and immediately rebuild your image, it's possible to iteratively explore various aspects of your app, such as by adding log lines that print variable values.

But to help you solve the problem _as fast as possible_, `ko` supports debugging your Go app with [delve](https://github.com/go-delve/delve).

To use this feature, just add the `--debug` flag to your `ko build` command. This adjusts how the image is built:

- It installs `delve` in the image (in addition to your own app).
- It sets the image's `ENTRYPOINT` to a `delve exec ...` command that runs the Go app in debug-mode, listening on port `40000` for a debugger client.
- It ensures your compiled Go app includes debug symbols needed to enable debugging.

**Note:** This feature is geared toward development workflows. It **should not** be used in production.

### How it works

Build the image using the debug feature.

```plaintext
ko build . --debug
```

You can also pass `--debug` to `ko apply` or `ko resolve` so Kubernetes manifests point at a debug-enabled image.

Run the container, ensuring that the debug port (`40000`) is exposed to allow clients to connect to it.

```plaintext
docker run -p 40000:40000 <img>
```

On Kubernetes, publish port `40000` from the Pod and forward it locally:

```plaintext
kubectl port-forward deploy/<your-deployment> 40000:40000
```

The process waits for a debugger client before your program runs. Connect with the [delve](https://github.com/go-delve/delve) CLI:

```plaintext
dlv connect 127.0.0.1:40000
```

Or use an editor that speaks the Delve remote protocol (for example VS Code `connect` / `request: attach` to `127.0.0.1:40000`).

### Tips

- Prefer a base image that includes a shell and basic tools if you also need to `kubectl exec` into the container while debugging.
- Debug builds keep symbols and skip optimizations that would make stepping through code harder. Do not ship `--debug` images to production.
- Remote debugging tools that inject agents into Pods (for example older workflows based on Squash) still need a writable container filesystem and often a non-distroless base image. The built-in `--debug` path above is the supported ko-native approach.
