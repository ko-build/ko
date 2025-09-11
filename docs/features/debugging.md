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

Run the container, ensuring that the debug port (`40000`) is exposed to allow clients to connect to it.

```plaintext
docker run -p 40000:40000 <img>
```

This sets up your app to be waiting to run the command you've specified. All that's needed now is to connect your debugger client to the running container!

By default, the application will not start until a debugger has connected and issued the `continue` command. This is required in order to be able to set breakpoints for code that is executed during start-up. If you want the application to start running without waiting for a debugger, use the `--debug-continue` parameter:

```plaintext
ko build . --debug --debug-continue
```
