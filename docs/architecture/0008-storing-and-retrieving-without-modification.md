# 8. SBOM storage/retrieval without modification

Date: 2024-10-16

## Status

Accepted

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

- Related Updates:
  - A protobom [feature](https://github.com/protobom/protobom/issues/213) to capture the origin format during deserialization.
  - Stores entire format string to validate format/encoding/version, ex: `"application/vnd.cyclonedx+xml;version=1.5"`

### Retrieval

TLDR:

- **Default Behavior:** If a document has been unaltered in the cache, bomctl will always output original document content if
desired format matches origin format.
- **Flag Behavior:** If the command contains `--original` flag to export or push cmd, the original document content will be used regardless if
the document had been altered in the cache or a different format is requested.
  - Create a special `original` format type that triggers this behavior

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
  - All modified documents will be exported/pushed in their modified state in whichever format designated by the command.
  - **Exception:** if the command contains the `--original` flag, all documents will be exported as their original content
    in their original format.

| Modified | Format Matches | `--original` | Output           |
|----------|----------------|--------------|------------------|
| False    | True           | True         | Origin Content   |
| False    | True           | False        | Origin Content   |
| False    | False          | True         | Origin Content   |
| False    | False          | False        | Updated Format   |
| True     | True           | True         | Origin Content   |
| True     | True           | False        | Modified Content |
| True     | False          | True         | Origin Content   |
| True     | False          | False        | Modified Content |

Questions:

- If a cyclonedx xml document is imported, and user requests an export of a cyclonedx json document which is unmodified, would we:
  - Export an xml document?
  - Internally convert origin content to json and export?
  - Output bomctl document as json?
    - Would this be lossless since it's the same format?

## Consequences

- Users will be able to export sbom documents in original format
  - To preserve signature validity.
  - To archive/catalog original documents acquired from the supply chain.
- Users may want to output to original format while including changes made in bomctl, which may not be lossless.
- Users may be surprised that `--original` ignores any modifications made in bomctl and will not be reflected in exported document.
  - Although the default functionality and behavior listed above seems intuitive.
