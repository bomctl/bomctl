# bomctl

__bomctl__ is format-agnostic Software Bill of Materials (SBOM) tooling, which is intended to bridge the gap between SBOM generation and SBOM analysis tools. It focuses on supporting more complex SBOM operations by being opinionated on only supporting the [NTIA minimum fields](https://www.ntia.doc.gov/files/ntia/publications/sbom_minimum_elements_report.pdf) or other fields supported by [protobom](https://github.com/bom-squad/protobom).

> [!NOTE]
> This is an experimental project under active development. We'd love feedback on the concept, scope, and architecture!

## Features

- Work with multiple SBOMs in tree structures (through external references)
- Fetch and push SBOMs using HTTPS, OCI, and GIT protocols
- Leverage a `.netrc` file to handle authentication
- Manipulate SBOMs with commands like `diff`, `split`, and `redact`
- Manage SBOMs using a persistent database cache
- Interface with OpenSSF projects and services like [GUAC](https://guac.sh/) and [Sigstore](https://www.sigstore.dev/)

## Join our Community

- [#bomctl on OpenSSF Slack](https://openssf.slack.com/archives/C06ED5VB81W)
- [OpenSSF Security Tooling Working Group Meeting](https://zoom-lfx.platform.linuxfoundation.org/meeting/94897563315?password=7f03d8e7-7bc9-454e-95bd-6e1e09cb3b0b) - Every two weeks on Friday, 8am Pacific
- [SBOM Tooling Working Meeting](https://zoom-lfx.platform.linuxfoundation.org/meeting/92103679564?password=c351279a-5cec-44a4-ab5b-e4342da0e43f) - Every Monday, 2pm Pacific

## Commands

### Fetch (Implemented)

Ability to retrieve an SBOM via several protocols:

- HTTP/S
- Git

and from various locations:

- Local Filesystem
- OCI

This includes recursive loading of external references in an SBOM to other SBOMs and placing them into the persistent cache. If SBOMs are access controlled, a user's [.netrc](https://www.gnu.org/software/inetutils/manual/html_node/The-_002enetrc-file.html) file to authenticate.

### Diff (Planned)

TBD

### Lint (Planned)

TBD

### List (Planned)

TBD

### Merge (Planned)

TBD

### Push (Planned)

TBD

### Redact (Planned)

TBD

### Split (Planned)

TBD

### Trim (Planned)

TBD
