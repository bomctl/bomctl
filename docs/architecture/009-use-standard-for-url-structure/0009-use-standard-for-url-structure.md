
# 9. Use Standard for URL format

Date: 2024-11-17

## Status

Proposed

## Context

The current command structure requires a 'url' to specify the location bomctl needs to fetch an sbom from or
push an sbom (or collection of sboms) to. The resposibility of parsing is left up to the individual clients,
which as we add more clients the number of individual places that the url gets parsed increases as well as the
number of potential corner cases or duplicate matches. In addition the needs of client apis may vary wildly and
force the addition of fields into the url structure to gather the requisite information. All this while the url
may become increasingly unwieldy for users to accurately enter or decipher which leads to mistakes in entry,
which increases frustration with bomctl.

The current url parsing scheme is an amalgam of various schemes, like those used for ansible and pip support as
well as traditional URL structure. The intention with the original url schema support as many forms of expression
as possible and stay as close to valid URL structure as possible. Using a home grown, non standard url structure
makes it harder for users to learn and harder for us to maintain long term. It also doesn't help that the current
psuedo url structure is not documented anywhere an thus bomctl is hard to pick up and use quickly.

Some examples of current url structure in their respective commands:

- HTTP Client
  - `example.acme.com`
  - `https://github.com/bomctl/bomctl/releases/download/v0.4.1/bomctl_0.4.1_darwin_amd64.tar.gz.spdx.json`
- Git Client
  - `git+https://git@github.com/bomctl/bomctl.git@main#sbom.cdx.json`
  - `ssh://git@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json`
  - `git://username:password@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json`
  - `git@github.com:bomctl/bomctl.git@main#sbom.cdx.json`
  - `https://github.com/bomctl/bomctl.git@main#path/to/sbom.cdx.json`
- OCI Client
  - `oci://username@registry.acme.com:12345/example/image:1.2.3`
  - `registry.acme.com/example/image:1.2.3`

The advantages of using an industry standard schema that fits our needs has two main advantages:

- There are existing tools and libraries that we can leverage to parse the command into composite parts
- Users may already have experience with the schema and thus will have an easier time getting their work done via bomctl

This ADR proposes a two part change:

- Unify the parsing of command strings into a single top level parsing function that determines the proper client
to use or fails with a message to the user.
- Switch from a custom syntax to an industry standard to simplify parsing and use.

### Unify Command Target Parsing

#### Current parsing structure

Main issues being:

- Managing N # of regular expressions to only capture items for that client and nothing else
- The iterative nature means a cmd intended for client N+3 could be captured by a mistake in the parsing function for N+1.

```mermaid
flowchart TD
  A[cmd] --> B(client.New)
  B --> C(For Each Client)
  C --> D(Client_N)
  D --> E(Client_N.Parse)
  E --> F(Match?)
  F -- No --> C
  F -- Yes --> G(Process cmd)
```

#### Proposed Parsing Structure

Improvements being:

- All parsing is done in one function
- Assuming no error in parsing, correct client will be selected
- No iteration to potentially be sent to incorrect client (Although that possibility is still there)

Drawbacks:

- Parsing function would be key point of failure
- Would require extensive testing suite
- May prove difficult to make additions

```mermaid
flowchart TD
  A[cmd] --> B(client.New)
  B --> C(client.Parse)
  C --> D(Error?)
  D -- Yes --> B
  D -- No --> E(Client_N)
  E --> F(Process cmd)
```

### Move to Industry Standard syntax for cmd information

Options (Please suggest any that may be a good fit)

- Package URL [(PURL)](https://github.com/package-url/purl-spec)
- net/url Go URL [Library](https://pkg.go.dev/net/url@go1.23.3). AKA Standard URL Structure

#### Package URL

Detailed writeup on using Package URL [here](0009-supporting-packageurl.md)

#### net/url

Detailed writeup on using URL Schema [here](0009-supporting-standard-url.md)

### Comparison of Example location strings

- HTTP Client
  - Current: `example.acme.com`
    - Purl: `pkg:generic/acme?download_url=example.acme.com`
    - URL:
  - Current: `https://github.com/bomctl/bomctl/releases/download/v0.4.1/bomctl_0.4.1_darwin_amd64.tar.gz.spdx.json`
    - Purl: `pkg:generic/bomctl@0.4.1?download_url=https://github.com/bomctl/bomctl/releases/download/v0.4.1/bomctl_0.4.1_darwin_amd64.tar.gz.spdx.json`
    - URL:
    - Notes: If we want the future github client to take this:
      - `pkg:github/bomctl/bomctl@0.4.1#bomctl_0.4.1_darwin_amd64.tar.gz.spdx.json`
- Git Client (Although these examples could use the future github client with purls also)
  - Current: `git+https://git@github.com/bomctl/bomctl.git@main#sbom.cdx.json`
    - Purl: `pkg:git/bomct/bomctl@main?user=git#sbom.cdx.json`
    - URL:
  - Current: `ssh://git@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json`
    - Purl: `pkg:git/bomct/bomctl@main?user=git&port=12345&scheme=ssh#sbom.cdx.json`
    - URL:
  - Current: `git://username:password@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json`
    - Purl: `pkg:git/bomct/bomctl@main?user=git&password=password&port=12345&scheme=git#sbom.cdx.json`
    - URL:
  - Current: `git@github.com:bomctl/bomctl.git@main#sbom.cdx.json`
    - Purl: `pkg:git/bomct/bomctl@main?user=git#sbom.cdx.json`
    - URL:
  - Current: `https://github.com/bomctl/bomctl.git@main#path/to/sbom.cdx.json`
    - Purl: `pkg:git/bomct/bomctl@main?user=git#path/to/sbom.cdx.json`
    - URL:
- OCI Client
  - Current: `oci://username@registry.acme.com:12345/example/image:1.2.3`
    - Purl: `pkg:oci/example/image@1.2.3?repository_url=registry.acme.com&username=username&port=12345`
    - URL:
  - Current: `registry.acme.com/example/image:1.2.3`
    - Purl: `pkg:oci/example/image@1.2.3?repository_url=registry.acme.com`
    - URL:

## Decision

## Consequences
