<!--
This is a template for [Documenting Architecture Decisions - Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).

You can use [adr-tools](https://github.com/npryce/adr-tools) for managing the ADR files.

In each ADR file, write the following sections.
-->
# 7. SBOM Diff

Date: 2024-07-30

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
Support the ability to see the difference between two SBOMs.

## Decision
<!--
This section describes our response to these forces. It is stated in full sentences, with active voice. "We will …"
-->
Seeing where two SBOMs differ and highlighting potentially problematic changes such as:

- a dependency that is slightly renamed and pointing to a different url
- a new dependency that is pointing to a new/unknown package
- a dependency whose version is bumped down

 Some tools do that do that on the package lock level since that sort of supply chain attack is starting to get a bit more prevalent.

Are there other useful things we should highlight for the user?

## Consequences
<!--
This section describes the resulting context, after applying the decision. All consequences should be listed here, not
just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them
affect the team and project in the future.
-->
