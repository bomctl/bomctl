# 8. SBOM storage/retrieval without modification

Date: 2024-10-16

## Status

Proposed

## Context

Documents imported into/exported by bomctl may have signatures associated with them, and we should support a way for users
to import an sbom and export the sbom in its' original form so that previous signatures can be validated. This is especially
true for documents that were not modified by the user while stored in the cache and are exported in the original format
they were imported in.

## Decision

### Storage

The storage in original format portion of this ADR is already implemented. Current behavior for all sbom documents added to the
local cache is to store the original document bytes as a unique annotation. Similarly, the original format of the sbom is stored
as a unique annotation at the time that its added to the cache.

- Outstanding Updates:
  - A protobom [feature](https://github.com/protobom/protobom/issues/213) to capture the origin format during deserialization.

- Optional improvements:
  - Store format version as an annotation
    - Do we want to be as granular as making sure the format version matches desired output format version before using source?

### Retrieval

TLDR:

- **Default Behavior:** If a document has been unaltered in the cache, bomctl will always output original document content if
desired format matches origin format.
- **Flag Behavior:** If the command contains `--original` flag to export or push cmd, the original document content will be used regardless if
the document had been altered in the cache or a different format is requested. (maybe we have a 'original' format type?)

Some Scenarios:
Context: A user imported multiple documents and then exports/pushes them with varying document states.

- All Modified Documents:
  - Modified Documents will be exported/pushed in their modified state in whichever format designated by the command.
  - **Exception:** if the command contains the `--original` flag, all documents will be exported as their original content
    in their original format.
- All Original Documents:
  - All original (unmodified) documents will be exported as their original content if the export format value matches the
    original format of the document, else will be exported as the format specified by the command.
  - **Exception:** if the command contains the `--original` flag, all documents will be exported as their original content
    in their original format.
- Mixed State: (Some modified, some unmodified)
  - All original (unmodified) documents will be exported as their original content if the export format value matches the
    original format of the document, else will be exported as the format specified by the command.
  - **Exception:** if the command contains the `--original` flag, all documents will be exported as their original content
    in their original format.

Questions:

- If a cyclonedx 1.5 document is imported, and user requests an export of a cyclonedx 1.6 formatted document which is unmodified, would we:
  - Export a 1.5 doc? (1.6 is a super set of 1.5 so it's technically valid), but:
    - Users may be confused to find a 1.5 document when they requested 1.6.
    - This also may have problems dopwn the line when using said document, if a tool is expecting a 1.6 (or doesn't support 1.5)
  - Implicitly convert to 1.6 and export as requested?
    - Causes the origin content to be exported in a slimmer window
    - Users would need to be very specific to get original content without `--original` flag
    - Implies the need to track/annotate origin format version

- If a cyclonedx xml document is imported, and user requests an export of a cyclonedx json document which is unmodified, would we:
  - Export an xml document?
  - Internally convert origin content to json and export?
    - Implies the need to track/annotate origin encoding
  - Output bomctl document as json?
    - Would this be lossless?
    - Implies the need to track/annotate origin encoding

## Consequences

- Users will be able to export sbom documents in original format
  - To preserve signature validity.
  - To archive/catalog original documents acquired from the supply chain.
- Users may want to output to original format while including changes made in bomctl, which won't happen with above solution.
- Users may be surprised that `--original` ignores and modifications made in bomctl and will not be reflected in exported document.
  - Although the default functionality and behavior listed above seems intuitive.
