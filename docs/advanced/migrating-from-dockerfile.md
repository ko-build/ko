# Migrating from Dockerfile

If your `Dockerfile` looks like either of the examples in the [official tutorial for writing a Dockerfile to containerize a Go application](https://docs.docker.com/language/golang/build-images/), you can easily migrate to use `ko` instead.

Let's review the best practice multi-stage Dockerfile in that tutorial first:

```plaintext
# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.16-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /docker-gs-ping

##
## Deploy
##
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /docker-gs-ping /docker-gs-ping

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/docker-gs-ping"]
```

This `Dockerfile`:

1. pulls the `golang:1.16` image
1. `COPY`s your local source into the container environment (`COPY`ing `go.mod` and `go.sum` first and running `go mod download`, to cache dependencies in the container environment)
1. `RUN`s `go build` on your source, inside the container, to produce an executable
1. `COPY`s the executable built in the previous step into a new image, on top of a minimal [distroless](https://github.com/GoogleContainerTools/distroless) base image.

The result is a Go application built on a minimal base image, with an optimally cached build sequence.

After running `docker build` on this `Dockerfile`, don't forget to push that image to the registry so you can deploy it.

---

## Migrating to `ko`

If your Go source is laid out as described in the tutorial, and you've [installed](../../install) and [set up your environment](../../get-started), you can simply run `ko build ./` to build and push the container image to your registry.

You're done. You can delete your `Dockerfile` and uninstall `docker`.

`ko` takes advantage of your local [Go build cache](../../features/build-cache) without needing to be told to, and it sets the `ENTRYPOINT` and uses a nonroot distroless base image by default.

To build a multi-arch image, simply add `--platform=all`.
Compare this to the [equivalent Docker instructions](https://docs.docker.com/desktop/multi-arch/).

