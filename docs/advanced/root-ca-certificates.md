# Root CA Certificates

To install a [root certificate](https://en.wikipedia.org/wiki/Root_certificate) into your container built using `ko`, you can use one of the following methods.

## Custom Base Image

New root certificates can be [installed into a custom image](https://stackoverflow.com/questions/42292444/how-do-i-add-a-ca-root-certificate-inside-a-docker-image) using standard OS packages. Then, this custom image can be used [to override the base image for `ko`](https://ko.build/configuration/#overriding-base-images). Once the Go application container image is built using `ko` with the custom base image, the root certificates installed on the base image will be trusted by the Go application.

### Example

1. Make a custom container image with your new root certificates
```dockerfile
# Dockerfile
FROM alpine

RUN apk update
RUN apk add ca-certificates

ADD new-root-ca.crt /usr/local/share/ca-certificates/new-root-ca.crt
RUN chmod 644 /usr/local/share/ca-certificates/new-root-ca.crt
RUN update-ca-certificates
```

2. Build and push the custom container image to a container registry
```sh
docker build . -t docker.io/ko-build/image-with-new-root-certs
docker push docker.io/ko-build/image-with-new-root-certs
```

3. Configure `ko` to [override the default base image](https://ko.build/configuration/#overriding-base-images) with the custom image
```yaml
# .ko.yaml
defaultBaseImage: docker.io/ko-build/image-with-new-root-certs
```

    **OR**
```sh
export KO_DEFAULTBASEIMAGE=docker.io/ko-build/image-with-new-root-certs
```

4. Build the Go app container image with `ko`
```sh
ko build .
```

## Static Assets
Alternatively, root certificates can be installed into the Go application container image using a combination of [`ko` static assets](https://ko.build/features/static-assets/) and [overriding the default system location for SSL certificates](https://pkg.go.dev/crypto/x509#SystemCertPool).

Using `ko`'s support for static assets, root certificates can be stored in the `<importpath>/kodata` directory (either checked into the repository, or injected dynamically by a CI pipeline). After running `ko build`, the certificate files are then bundled into the built image at the path `$KO_DATA_PATH`.

To enable the Go application to trust the bundled certificate(s), the container runtime or orchestrator (Docker, Kubernetes, etc) must set the environment variable `SSL_CERT_DIR` to the same value as `KO_DATA_PATH`. Go [uses `SSL_CERT_DIR` to determine the directory to check for SSL certificate files](https://go.dev/src/crypto/x509/root_unix.go). Once this variable is set, the Go application will trust the bundled root certificates in `$KO_DATA_PATH`.

### Example

1. Copy the root certificate(s) to the `<importpath>/kodata/` directory
```sh
# $(pwd) assumed to be at <importpath> for this example
mkdir -p kodata
cp $CERT_FILE_DIR/*.crt kodata/
```

2. Build the Go application container image
```sh
KO_DOCKER_REPO=docker.io/ko-build/static-assets-certs ko build .
```

3. Run the Go application container image with `SSL_CERT_DIR` equal to `/var/run/ko` (the default value for `$KO_DATA_PATH`)
```sh
docker run -e SSL_CERT_DIR=/var/run/ko docker.io/ko-build/static-assets-certs
```

A functional client-server example for this can be seen [here](https://github.com/kosamson/ko-private-ca-test).
