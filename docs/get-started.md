# Get Started

## Setup

First, [install `ko`](../install).

### Authenticate

`ko` depends on the authentication configured in your Docker config (typically `~/.docker/config.json`).

✨ **If you can push an image with `docker push`, you are already authenticated for `ko`!** ✨

Since `ko` doesn't require `docker`, `ko login` also provides a surface for logging in to a container image registry with a username and password, similar to [`docker login`](https://docs.docker.com/engine/reference/commandline/login/).

Additionally, even if auth is not configured in the Docker config, `ko` includes built-in support for authenticating to the following container registries using credentials configured in the environment:

- Google Container Registry and Artifact Registry, using [Application Default Credentials](https://cloud.google.com/docs/authentication/production) or auth configured in `gcloud`.
- Amazon Elastic Container Registry, using [AWS credentials](https://github.com/awslabs/amazon-ecr-credential-helper/#aws-credentials)
- Azure Container Registry, using [environment variables](https://github.com/chrismellard/docker-credential-acr-env/)
- GitHub Container Registry, using the `GITHUB_TOKEN` environment variable

### Choose Destination

`ko` depends on an environment variable, `KO_DOCKER_REPO`, to identify where it should push images that it builds. Typically this will be a remote registry, e.g.:

- `KO_DOCKER_REPO=gcr.io/my-project`, or
- `KO_DOCKER_REPO=ghcr.io/my-org/my-repo`, or
- `KO_DOCKER_REPO=my-dockerhub-user`

## Build an Image

`ko build ./cmd/app` builds and pushes a container image, and prints the resulting image digest to stdout.

In this example, `./cmd/app` must be a `package main` that defines `func main()`.

```plaintext
$ ko build ./cmd/app
...
registry.example.com/my-project/app-099ba5bcefdead87f92606265fb99ac0@sha256:6e398316742b7aa4a93161dce4a23bc5c545700b862b43347b941000b112ec3e
```

> 💡 **Note**: Prior to v0.10, the command was called `ko publish` -- this is equivalent to `ko build`, and both commands will work and do the same thing.

The executable binary that was built from `./cmd/app` is available in the image at `/ko-app/app` -- the binary name matches the base import path name -- and that binary is the image's entrypoint.

