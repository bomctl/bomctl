# Updating Primary Component Metadata

## Description

A common usecase is for a software producer to create an SBOM using a tool like trivy, syft, or other tool in a CI/CD pipeline. The generated SBOM often lacks the metadata for license, copyright, descriptions, and external references like documentation.

## Desire Input

- [An SBOM of any format](metadata.cyclonedx.json) that only contains the metadata that needs to be updated and associated
identifier to update.
- A generated SBOM of any format that comes from syft, trivy, or other tool

## Desired Output

- SBOMs in SPDX and CDX that have the generated SBOM and updated metadata

## Operations

- Load SBOMS into `bomctl`

```shell
bomctl import --alias metadata metadata.cyclonedx.json
bomctl import --alias syft syft.cdx.json
```

- Merge SBOMS

```shell
bomctl merge --alias merged metadata syft
```

- Output merged SBOM into multiple formats

```shell
bomctl export --encoding json --format cyclonedx-1.5 bomctl.tar.gz.cdx.json
bomctl export --encoding json --format spdx-2.3 bomctl.tar.gz.spdx.json
```
