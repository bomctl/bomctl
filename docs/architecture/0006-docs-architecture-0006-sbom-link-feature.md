<!--
This is a template for [Documenting Architecture Decisions - Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).

You can use [adr-tools](https://github.com/npryce/adr-tools) for managing the ADR files.

In each ADR file, write the following sections.
-->
# 6. SBOM Linking

Date: 2024-09-18

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

<!-- What is the issue that we're seeing that is motivating this decision or change? -->

### SBOM Document Linking (`link` command)

Bomctl needs the ability to create external references between documents. A use-case for this
feature could be a user that containerizes an application and then needs an SBOM for the new
container, with external references to the SBOMs of the container image and the application
binaries.

Another feature to discuss is the ability to express relationships between the components in the
SBOM.

### Implications

- A user creates two SBOMs __locally__, and then uses `bomctl` to link one of the SBOMs to a component
in the other document.
  - What should the `externalRef` from the child SBOM to the parent SBOM look like?
  - When these documents are pushed to another service (OCI registry, github release, etc), should these
    `externalRef` be resolved and updated to be their "pushed" location?
  - When a new link is created is a new parent document created? What about the parent's parent document?
- A user creates an SBOM, and then uses `bomctl` to link to an external document.
  - What should the `externalRef` from the child SBOM to the parent SBOM look like?
  - When these documents are pushed to another service (OCI registry, github release, etc), should these
    `externalRef` be resolved and updated to be their "pushed" location?
  - When a new link is created is a new parent document created? What about the parent's parent document?
- A user creates an SBOM, and then uses `bomctl` to link to a document that was previously loaded from
  an external document?
  - What should the `externalRef` from the child SBOM to the parent SBOM look like?
  - When these documents are pushed to another service (OCI registry, github release, etc), should these
    `externalRef` be resolved and updated to be their "pushed" location?
  - When a new link is created is a new parent document created? What about the parent's parent document?

### Other Considerations

- We have plans to sign SBOMs, when should this happen?

## Decision
<!--
This section describes our response to these forces. It is stated in full sentences, with active voice. "We will …"
-->

<!-- What is the change that we're proposing and/or doing? -->

__TBD.__

## Consequences
<!--
This section describes the resulting context, after applying the decision. All consequences should be listed here, not
just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them
affect the team and project in the future.
-->

<!-- What becomes easier or more difficult to do because of this change? -->

:heavy_check_mark: __Cross-referencing and tracking components will become easier__. By allowing
external references and expressing relationships between components, tracking and understanding the
connections between various parts of a system will become more manageable. This will enable more
efficient analysis, such as impact assessment when making changes or upgrades.

:heavy_check_mark: __Vulnerability scanning will be more thorough__. With external references, it
will be easier to correlate SBOM components with known vulnerabilities in external repositories,
improving the overall security of the system.

:x: __Data validation and consistency will become harder__. With external references, ensuring data
consistency and validating the accuracy of information from external sources could become more
challenging. Implementing robust validation and error-handling mechanisms will be crucial to
maintain data integrity.
