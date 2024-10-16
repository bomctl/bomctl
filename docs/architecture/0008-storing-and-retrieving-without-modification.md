# 8. SBOM storage/retrieval without modification

Date: 2024-10-16

## Status

Proposed

## Context
<!--
This section describes the forces at play, including technological, political, social, and project local. These forces
are probably in tension, and should be called out as such. The language in this section is value-neutral. It is simply
describing facts.
-->
Documents imported into/exported by bomctl may have signatures associated with them, and we should support a way for users
to import an sbom and export the sbom in its' original form so that previous signatures can be validated. This is especially
true for documents that were not modified by the user while stored in the cache and are exported in the original format
they were imported in.

## Decision
<!--
This section describes our response to these forces. It is stated in full sentences, with active voice. "We will …"
-->

### Storage

The storage in original format portion of this ADR is already implemented. Current behavior for all sbom documents added to the
local cache is to store the original document bytes as a unique annotation. Similarly, the original format of the sbom is stored
as a unique annotation at the time that its added to the cache.

- Outstanding Updates:
  - a small protobom change to properly set the document format during sbomReader.parseStream

- Optional improvements:
  - Store format version as an annotation
    - Do we want to be as granular as making sure the format version matches desired output format version before using source?

### Retrieval

**Default Behavior:** If a document has been unaltered in the cache, bomctl will always output original document content if
desired format matches origin format.

**Flag Behavior:** If a user passes `--original` to export or push cmd, the original document content will be used regardless if
the document had been altered in the cache or a different format is requested. (maybe we have a 'original' format type?)

Some Scenarios:

- All Internal Documents:
- All External Documents:
- Mixed Origin

## Consequences
<!--
This section describes the resulting context, after applying the decision. All consequences should be listed here, not
just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them
affect the team and project in the future.
-->
