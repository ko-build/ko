# SBOMs

A [Software Bill of Materials (SBOM)](https://en.wikipedia.org/wiki/Software_bill_of_materials) is a list of software components that a software artifact depends on.
Having a list of dependencies can be helpful in determining whether any vulnerable components were used to build the software artifact.

**From v0.9+, `ko` generates and uploads an SBOM for every image it produces by default.**

ko will generate an SBOM in the [SPDX](https://spdx.dev/) format by default, but you can select the [CycloneDX](https://cyclonedx.org/) format instead with the `--sbom=cyclonedx` flag. To disable SBOM generation, pass `--sbom=none`.

These SBOMs can be downloaded using the [`cosign download sbom`](https://github.com/sigstore/cosign/blob/main/doc/cosign_download_sbom.md) command.

