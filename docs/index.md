---
ko_meta: true
---

# Introduction

`ko` makes building Go container images easy, fast, and secure by default.

![Demo of ko build](./images/demo.png)

`ko` is a simple, fast container image builder for Go applications.

It's ideal for use cases where your image contains a single Go application without many dependencies on the OS base image (e.g., no cgo, no OS package dependencies).

`ko` builds images by executing `go build` on your local machine, and as such doesn't require `docker` to be installed.
This can make it a good fit for lightweight CI/CD use cases.

`ko` also includes support for simple YAML templating which makes it a powerful tool for [Kubernetes applications](./features/k8s).

---

> 🏃 [Install `ko`](./install) and [get started](./get-started)!

---

`ko` is used and loved by these open source projects:

- [Knative](https://knative.dev)
- [Tekton](https://tekton.dev)
- [Karpenter](https://karpenter.sh)
- [Sigstore](https://sigstore.dev)
- [Shipwright](https://shipwright.io)

[_Add your project here!_](https://github.com/imjasonh/ko.build/edit/main/docs/index.md)

