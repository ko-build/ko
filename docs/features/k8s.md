# Kubernetes Integration

You _could_ stop at just building and pushing images.

But, because building images is so _easy_ with `ko`, and because building with
`ko` only requires a string importpath to identify the image, we can integrate
this with YAML generation to make Kubernetes use cases much simpler.

## YAML Changes

Traditionally, you might have a Kubernetes deployment, defined in a YAML file,
that runs an image:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  replicas: 3
  ...
  template:
    spec:
      containers:
      - name: my-app
        image: registry.example.com/my-app:v1.2.3
```

...which you apply to your cluster with `kubectl apply`:

```plaintext
kubectl apply -f deployment.yaml
```

With `ko`, you can instead reference your Go binary by its importpath, prefixed
with `ko://`:

```yaml
    ...
    spec:
      containers:
      - name: my-app
        image: ko://github.com/my-user/my-repo/cmd/app
```

## `ko resolve`

With this small change, running `ko resolve -f deployment.yaml` will instruct
`ko` to:

1. scan the YAML file(s) for values with the `ko://` prefix,
2. for each unique `ko://`-prefixed string, execute `ko build <importpath>` to
   build and push an image,
3. replace `ko://`-prefixed string(s) in the input YAML with the fully-specified
   image reference of the built image(s), as above.
4. Print the resulting resolved YAML to stdout.

The result can be redirected to a file, to distribute to others:

```plaintext
ko resolve -f config/ > release.yaml
```

Taken together, `ko resolve` aims to make packaging, pushing, and referencing
container images an invisible implementation detail of your Kubernetes
deployment, and let you focus on writing code in Go.

## `ko apply`

To apply the resulting resolved YAML config, you can redirect the output of
`ko resolve` to `kubectl apply`:

```plaintext
ko resolve -f config/ | kubectl apply -f -
```

Since this is a relatively common use case, the same functionality is available
using `ko apply`:

```plaintext
ko apply -f config/
```

Also, any flags passed after `--` are passed to `kubectl apply` directly, for example to specify context and kubeconfig:
```
ko apply -f config -- --context=foo --kubeconfig=cfg.yaml
```

**NB:** This requires that `kubectl` is available.

## `ko delete`

To teardown resources applied using `ko apply`, you can run `ko delete`:

```plaintext
ko delete -f config/
```

This is purely a convenient alias for `kubectl delete`, and doesn't perform any
builds, or delete any previously built images.

