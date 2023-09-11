# AWS Lambda

`ko` can build images that can be deployed as AWS Lambda functions, using [Lambda's container support](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html).

For best results, use the [Go runtime interface client](https://docs.aws.amazon.com/lambda/latest/dg/go-image.html#go-image-clients) provided by the [`lambda` package](https://pkg.go.dev/github.com/aws/aws-lambda-go/lambda).

For example:

```go
package main

import (
    "fmt"
    "context"
    "github.com/aws/aws-lambda-go/lambda"
)

type Event struct {
    Name string `json:"name"`
    // TODO: add other request fields here.
}

func main() {
    lambda.Start(func(ctx context.Context, event Event) (string, error) {
        return fmt.Sprintf("Hello %s!", event.Name), nil
    }
}
```

See AWS's [documentation](https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html) for more information on writing Lambda functions in Go.

To deploy to Lambda, you must push to AWS Elastic Container Registry (ECR):

```sh
KO_DOCKER_REPO=[account-id].dkr.ecr.[region].amazonaws.com/my-repo
image=$(ko build ./cmd/app)
```

Then, create a Lambda function using the image in ECR:

```sh
aws lambda create-function \
  --function-name hello-world \
  --package-type Image \
  --code ImageUri=${image} \
  --role arn:aws:iam::[account-id]:role/lambda-ex
```

See AWS's [documentation](https://docs.aws.amazon.com/lambda/latest/dg/go-image.html) for more information on deploying Lambda functions using Go container images, including how to configure push access to ECR, and how to configure the IAM role for the function.

The base image that `ko` uses by default supports both x86 and Graviton2 architectures.

You can also use the [`ko` Terraform provider](./terraform.md) to build and deploy Lambda functions as part of your IaC workflow, using the [`aws_lambda_function` resource](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_function.html). See the [provider example](https://github.com/ko-build/terraform-provider-ko/tree/main/provider-examples/lambda) to get started.
