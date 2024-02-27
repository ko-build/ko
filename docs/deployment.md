# Deployment

_See [Kubernetes Integration](../features/k8s) for information about deploying to Kubernetes._

Because the output of `ko build` is an image reference, you can easily pass it to other tools that expect to take an image reference.

### [`docker run`](https://docs.docker.com/engine/reference/run/)

To run the container locally:

```plaintext
docker run -p 8080:8080 $(ko build ./cmd/app)
```

---

### [Google Cloud Run](https://cloud.google.com/run)

```plaintext
gcloud run deploy --image=$(ko build ./cmd/app)
```

> 💡 **Note:** The image must be pushed to [Google Container Registry](https://cloud.google.com/container-registry) or [Artifact Registry](https://cloud.google.com/artifact-registry).

---

###  [fly.io](https://fly.io)

```plaintext
flyctl launch --image=$(ko build ./cmd/app)
```

> 💡 **Note:** The image must be pushed to Fly.io's container registry at `registry.fly.io`, or if not, the image must be publicly available. When pushing to `registry.fly.io`, you must first log in with [`flyctl auth docker`](https://fly.io/docs/flyctl/auth-docker/).

---

### [AWS Lambda](https://aws.amazon.com/lambda/)

```plaintext
aws lambda update-function-code \
  --function-name=my-function-name \
  --image-uri=$(ko build ./cmd/app)
```

> 💡 **Note:** The image must be pushed to [ECR](https://aws.amazon.com/ecr/), based on the AWS provided base image, and use the [`aws-lambda-go`](https://github.com/aws/aws-lambda-go) framework.
See [official docs](https://docs.aws.amazon.com/lambda/latest/dg/go-image.html) for more information.

---

### [Azure Container Apps](https://azure.microsoft.com/services/container-apps/)

```plaintext
az containerapp update \
  --name my-container-app
  --resource-group my-resource-group
  --image $(ko build ./cmd/app)
```

> 💡 **Note:** The image must be pushed to [ACR](https://azure.microsoft.com/services/container-registry/) or other registry service.
See [official docs](https://docs.microsoft.com/azure/container-apps/) for more information.

