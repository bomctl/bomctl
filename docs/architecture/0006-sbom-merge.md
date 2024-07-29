<!--
This is a template for [Documenting Architecture Decisions - Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).

You can use [adr-tools](https://github.com/npryce/adr-tools) for managing the ADR files.

In each ADR file, write the following sections.
-->
# 6. 0006-sbom-merge.md

Date: 2024-07-29

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
SBOM merge capability should be supported.

## Decision
<!--
This section describes our response to these forces. It is stated in full sentences, with active voice. "We will …"
-->
- Ability to take one or more SBOMs in different formats into a singular sbom
  - Merge components that may match on different identifiers (purls, hash, etc …)
    - Merge component dependencies
    - Merge component properties

- Ability to take a flag to “flatten” any externally referenced SBOMs
  - Any nested components should be moved to "top level" components and the depends_on section updated to show the dependency.

## Consequences
<!--
This section describes the resulting context, after applying the decision. All consequences should be listed here, not
just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them
affect the team and project in the future.
-->

These changes allow end users to easily consolidate multiple SBOMs from a variety of different sources into one while still being agnostic of the SBOM format being used.
