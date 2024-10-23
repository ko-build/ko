# Linux Capabilities

In Linux, capabilities are a way to selectively grant privileges to a running process.

Docker provides `--cap-add` and `--cap-drop`
[run options](https://docs.docker.com/engine/containers/run/#runtime-privilege-and-linux-capabilities)
to tweak container capabilities, e.g:

```
docker run --cap-add bpf hello-world
```

If container runs as a non-root user,
capabilities are narrowed by intersecting with *file* capabilities of the
application binary. When building images with a Dockerfile, one
typically uses `setcap` tool to modify file capabilities, e.g:
`setcap FILE bpf=ep`.

To set file capabilities with `ko`, specify `linux_capabilities`
in builds configuration section in your `.ko.yaml`. Use `setcap` syntax:

```yaml
builds:
- id: caps
  linux_capabilities: bpf=ep
```
## Alternative spelling

```yaml
builds:
- id: caps
  linux_capabilities:
  - cap1
  - cap2
  - cap3
```

A list of capability names is equivalent to `cap1,cap2,cap3=p`.

## Improving UX in capability-reliant apps

A capability can be *permitted* (`=p`), or both *permitted* and *effective* (`=ep`).
Effective capabilities are used for permission checks.
A program can promote permitted capability to effective when needed.

```yaml
builds:
- id: caps
  linux_capabilities: bpf,perfmon,net_admin=ep
```

Initially, `=ep` might look like a good idea.
There's no need to explicitly promote *permitted* capabilities.
Application takes advantage of *effective* capabilities right away.

There is a catch though.

```
$ docker run --cap-add bpf ko.local/caps.test-4b8f7bca75c467b3d2803e1c087a3287
exec /ko-app/caps.test: operation not permitted
```

When run options request fewer capabilities than specified in file capabilities,
container fails to start. It is hard to tell what went wrong since
`operation not permitted` is a generic error. (Docker is unlikely to improve diagnostics
for this failure case since the check is implemented in Linux kernel.)


We suggest to use `=p` instead. This option puts application in charge of verifying and
promoting permitted capabilities to effective. It can produce much better diagnostics:

```
$ docker run --cap-add bpf ko.local/caps.test-4b8f7bca75c467b3d2803e1c087a3287
current capabilities: cap_bpf=p
activating capabilities: cap_net_admin,cap_perfmon,cap_bpf=ep: operation not permitted
```

[Sample code](https://go.dev/play/p/uPMzyotkNHg).
