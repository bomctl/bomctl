<!--
This is a template for [Documenting Architecture Decisions - Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).

You can use [adr-tools](https://github.com/npryce/adr-tools) for managing the ADR files.

In each ADR file, write the following sections.
-->
# 2. Import subcommand

Date: 2024-07-09

## Status
<!--
A decision may be "proposed" if the project stakeholders haven't agreed with it yet, or "accepted" once it is agreed.
If a later ADR changes or reverses a decision, it may be marked as "deprecated" or "superseded" with a reference to
its replacement.
-->
Proposed

## Context
<!--
This section describes the forces at play, including technological, political, social, and project local. These forces
are probably in tension, and should be called out as such. The language in this section is value-neutral. It is simply
describing facts.
-->
There is currently no capability of importing SBOM or protobom data from either `stdin` or local filesystem path(s).

## Decision
<!--
This section describes our response to these forces. It is stated in full sentences, with active voice. "We will …"
-->
Introduce an `import` command that will accept one of the following options as input:

- stream of bytes piped from `stdin`
- path to a local file or files as optional positional arguments
  - alternatively, input files could could be specified with an explicit flag, such as `--input`, `--file`, `--path`, etc.

The supported input types will be:

- CycloneDX SBOM
- SPDX SBOM
- Protobom `Document` (such as the serialized protocol buffer storage format used by the Protobom `FileSystemBackend`)

The `fetch` command will now be able to simply fetch raw bytes data and leverage the new `import` logic to store.

## Consequences
<!--
This section describes the resulting context, after applying the decision. All consequences should be listed here, not
just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them
affect the team and project in the future.
-->
### Integration between CLI tools

Promotes integration with other CLI tools by accepting their piped output.

For example, this could enable usage patterns such as:

```shell
curl --silent --url https://acme.example.com/sbom.cdx.json | bomctl import
```

### Input flexibility

Presents additional input options for users that may feel more natural.

### Feature parity

Adds a counterpart to the `export` command for feature parity and completeness.
