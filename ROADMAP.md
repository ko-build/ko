# `ko` Project Roadmap

_Last updated October 2022_

- Foster a community of contributors and users
  - give talks, do outreach, expand the pool of contributors
  - identify projects that could benefit from using `ko`, and help onboard them
  - publish case studies from successful migrations

- Integrate [sigstore](https://sigstore.dev) for built artifacts
  - attach signed SBOMs
  - attach signed provenance attestations
  - support the OCI referrers API and [fallback tag scheme](https://github.com/opencontainers/distribution-spec/blob/main/spec.md#referrers-tag-schema)
  - integrate with CI workload identity (e.g., GitHub OIDC) to keylessly sign artifacts

- Faster builds
  - identify unnecessary work and avoid it when possible

- Ecosystem integrations
  - support Terraform provider, and potentially Pulumi and CDK, others
  - provide working examples of these integrations
