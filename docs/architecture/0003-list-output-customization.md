<!--
This is a template for [Documenting Architecture Decisions - Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).

You can use [adr-tools](https://github.com/npryce/adr-tools) for managing the ADR files.

In each ADR file, write the following sections.
-->
# 3. List output customization

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
This is an enhancement proposal to allow users to customize the output of `bomctl list` that lists
SBOM Documents by specifying a search expression, filter, or template.

## Decision
<!--
This section describes our response to these forces. It is stated in full sentences, with active voice. "We will …"
-->
Update `bomctl list` to accept an expression to customize the output table of SBOM documents.
This would be similar in functionality to the `--format` option used by various `docker` commands.

Either the existing use of positional argument will be retained, or an additional flag (such as
`--format` or `--query`) will be added.

The supported search/filter expression syntaxes will be:

- Common Expression Language (CEL)
- Go templates
- GraphQL

The expression would then be parsed to construct a SQL query or call to the `ent` backend to
retrieve and display the desired table(s)/column(s).

A separate `bomctl query` command will be used to query/filter nodes, `bomctl list` is reserved
for finding SBOM Documents.

@puerco recommended reading the store into CEL to simplify listing.

## Consequences
<!--
This section describes the resulting context, after applying the decision. All consequences should be listed here, not
just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them
affect the team and project in the future.
-->
This change would allow users greater control over inspecting the contents of their cached SBOMs.
