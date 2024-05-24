# bomctl

[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/bomctl/bomctl/badge)](https://securityscorecards.dev/viewer/?uri=github.com/bomctl/bomctl)
[![Go Report Card](https://goreportcard.com/badge/github.com/bomctl/bomctl)](https://goreportcard.com/report/github.com/bomctl/bomctl)
[![Go Reference](https://pkg.go.dev/badge/github.com/bomctl/bomctl.svg)](https://pkg.go.dev/github.com/bomctl/bomctl)
[![Slack](https://img.shields.io/badge/slack-openssf/bomctl-white.svg?logo=slack)](https://slack.openssf.org/#bomctl)

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

## Container Images

Container images for bomctl can be found on [Docker Hub](https://hub.docker.com/r/bomctl/bomctl).

``` shell
docker run bomctl/bomctl:latest --help
```

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

## Verifying Releases

Bomctl releases can be [found here](https://github.com/bomctl/bomctl/releases) and are signed
using keyless signing with cosign.

For each artifact there are two additional files:

- `${artifact}-keyless.sig` - signature
- `${artifact}-keyless.pem` - certificate

You can then verify this artifact with cosign using the signature and the certificate.

``` shell
cosign verify-blob --certificate ${artifact}-keyless.pem --signature ${artifact}-keyless.sig --certificate-oidc-issuer https://token.actions.githubusercontent.com --certificate-identity-regexp 'https://github\.com/bomctl/bomctl/\.github/.+'  ${artifact}
```

If the result is `Verified OK`, the verification is successful.

You can also look up the entry in the public Rekor instance using a sha256 hash.

``` shell
shasum -a 256 bomctl_SNAPSHOT-3f16bdb_checksums.txt |awk '{print $1}'
```

The printed `hash` can be used to look up the entry at <https://search.sigstore.dev/>.
