<!--
This is a template for [Documenting Architecture Decisions - Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).

You can use [adr-tools](https://github.com/npryce/adr-tools) for managing the ADR files.

In each ADR file, write the following sections.
-->
# 5. SBOM document aliases/tags

Date: 2024-07-11

## Status
<!--
A decision may be "proposed" if the project stakeholders haven't agreed with it yet, or "accepted" once it is agreed.
If a later ADR changes or reverses a decision, it may be marked as "deprecated" or "superseded" with a reference to
its replacement.
-->
Accepted

## Context
<!--
This section describes the forces at play, including technological, political, social, and project local. These forces
are probably in tension, and should be called out as such. The language in this section is value-neutral. It is simply
describing facts.
-->
A shorter form of referencing SBOM documents stored in `bomctl`'s database is needed, while also
adding adding annotations/labels for adding key/value pair metadata.

## Decision
<!--
This section describes our response to these forces. It is stated in full sentences, with active voice. "We will …"
-->
Add support for user assignment of annotations/labels per stored SBOM document:

- a list of labels (key/value pairs)
- the `alias` key is reserved and will be used as a unique short hand label for the SBOM document

These options could also be consolidated into a unified concept such as **labels**.

Users should be able to perform this action either at fetch/import or as a standalone command/operation.

## Consequences
<!--
This section describes the resulting context, after applying the decision. All consequences should be listed here, not
just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them
affect the team and project in the future.
-->

Implementing `alias` key could be limiting due to their 1:1 association.

Using the label methodology may provide greater flexibility in association between documents for
actions like storing and retrieving a group of SBOMs in a tree structure.
