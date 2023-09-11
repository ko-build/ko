# Terraform Provider

In addition to the CLI, `ko`'s functionality is also available as a Terraform provider.

This allows `ko` to be integrated with your Infrastructure-as-Code (IaC) workflows, and makes building your code a seamless part of your deployment process.

Using the Terraform provider is as simple as adding a `ko_build` resource to your Terraform configuration:

```hcl
// Require the `ko-build/ko` provider.
terraform {
  required_providers {
    ko = { source = "ko-build/ko" }
  }
}

// Configure the provider to push to your repo.
provider "ko" {
  repo = "example.registry/my-repo" // equivalent to KO_DOCKER_REPO
}

// Build your code.
resource "ko_build" "app" {
  importpath = "github.com/example/repo/cmd/app"
}

// TODO: use the `ko_build.app` resource elsewhere in your Terraform configuration.

// Report the build image's digest.
output "image" {
  value = ko_build.app.image
}
```

See the [`ko-build/ko` provider on the Terraform Registry](https://registry.terraform.io/providers/ko-build/ko/latest) for more information, and the [GitHub repo](https://github.com/ko-build/terraform-provider-ko) for more examples.
