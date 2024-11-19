
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

<details>
<summary>Package URL Breakdown</summary>

## Description

Package URL Specification (purl spec): A minimal specification for purl aka. a package "mostly universal" URL.

- [GitHub](https://github.com/package-url/purl-spec)

## Advantages

- Similar to the format in use currently in bomctl
  - Many of the fields bomctl is looking for exist within a purl
- Purls are used in sboms, making it a logical jump to use in bomctl
- Users may be familiar with the standard since it's used in sboms
- Some built in support for various clients (git, oci, github)
  - This would aid in determining which client should be used
- This would simplify commands for operations on boms:
  - Linking would require pointing to a component within an sbom, which would use purl spec

## Disadvantages

- Bare urls would need to be wrapped in a useless formatting, or go through additional processing to turn into a purl
  - `pkg:generic/bomctl@0.4.1?download_url=https://github.com/bomctl/bomctl/releases/download/v0.4.1/bomctl_0.4.1_darwin_amd64.tar.gz.spdx.json`
  - Or bare urls just get parsed and passed to http client. (make it a special case for usability)
- Does not support all prospective future clients (gitlab, dependency track, etc.)
  - Although gitlab is listed as a candidate for a potential future purl type
- Has support for many purl types we may never want to use (not necessarily a bad thing)
- Not sure this alleviates the problem of location strings in commands becoming and long and unwieldy with the use of qualifiers.
- Potentially requires users to know which client they'd want to use, which it'd be better to avoid
- Cannot express local files in this schema, may have to be handled as a generic string (as it is now)
  - Could also use generic pkg: `pkg:generic/sbom.cdx.json@v0.4.1?download_url=file:///path/to/sbom.cdx.json`

## Practicality/Usability

- Soft enforcement of `pkg` prefix
  - We could allow users to leave off the `pkg:` prefix from their cmd and handle it if it's not there.
- Component Mapping:
  - Most purl component parts map directly to information we are currently collecting
  - Namespace could handle nested groups (gitlab style: instance/group/group/.../repo) or github style (just org/repo)
  - Any information missing, could be stored as qualifiers, like in the below examples
- Qualifiers
  - Since purls do not directly handle many components that may be seen in bomctl, like username, port, scheme, etc.
    It's necessary to use qualifiers to add this information, which could make the cmd string very long.
  - If we're wanting authentication to go through .netrc, would we need to support cmds that have the username/password info in them?
    - Supporting username and password in commands does help simplify running in CICD environments (with proper precautions setup)
    - Maybe separate flags for this?

## Examples

- HTTP Client
  - `pkg:generic/acme?download_url=example.acme.com`
  - `pkg:generic/bomctl@0.4.1?download_url=https://github.com/bomctl/bomctl/releases/download/v0.4.1/bomctl_0.4.1_darwin_amd64.tar.gz.spdx.json`
    - If we would want the future github client to take this: `pkg:github/bomctl/bomctl@0.4.1#bomctl_0.4.1_darwin_amd64.tar.gz.spdx.json`
- Git  (Although all these examples could also have the prefix `pkg:github` and use the future github client)
  - `pkg:git/bomct/bomctl@main?user=git#sbom.cdx.json`
  - `pkg:git/bomct/bomctl@main?user=git&port=12345&scheme=ssh#sbom.cdx.json`
  - `pkg:git/bomct/bomctl@main?user=git&password=password&port=12345#sbom.cdx.json`
  - `pkg:git/bomct/bomctl@main?user=git#sbom.cdx.json`
  - `pkg:git/bomct/bomctl@main?user=git#path/to/sbom.cdx.json`
- OCI Client
  - `pkg:oci/example/image@1.2.3?repository_url=registry.acme.com&username=username&port=12345`
  - `pkg:oci/example/image@1.2.3?repository_url=registry.acme.com`

</details>

- net/url Go URL [Library](https://pkg.go.dev/net/url@go1.23.3). AKA Standard URL/URI Structure

<details>
<summary>net/url Breakdown</summary>

## Description

`[scheme:][//[userinfo@]host][/]path[?query][#fragment]`

- Scheme     : ex: http\[s\], file, ftp
- User       : username and password information (optional)
- Host       : host or host:port
- Path       : consisting of a sequence of path segments separated by a slash (/)
- RawQuery   : encoded query values, without '?'
- Fragment   : fragment for references, without '#'

Examples:

- `foo://example.com:8042/over/there?name=ferret#nose`
- `http://www.example.com/questions/3456/my-document`
- `ftp://ds.internic.net/internet-drafts/draft-ietf-uri-irl-fun-req-02.txt`
- `https://blog.hubspot.com/website/application-programming-interface-api`
- `http://www.ietf.org/rfc/rfc2396.txt`
- `file://this/one/over/here.json`

## Advantages

- Expressive standard that supports all fields currently in use by bomctl
- Very close to what we're already using in bomctl
- Built-in support for files `file://`
- Familiar syntax for users
- Abstracts client choice away from command texts
- Will require less translation, since sbom locations will probably be mostly urls
- Authentication info and port are embedded, not requiring extra qualifiers

## Disadvantages

- No support of different clients (git, gitlab, oci)
- Not sure this alleviates the problem of location strings in commands becoming and long and unwieldy with the use of queries/fragments.
- Command strings may be more ambiguous as to which client is needed,
leading to incorrect client chosen or may have to attempt and retry
with a different client upon failure.

## Practicality/Usablity

- Component Mapping:
  - Most uri component parts map directly to information we are currently collecting
  - Any information missing, could be stored as queries

## Examples

- HTTP Client
  - `https://example.acme.com`
  - `https://github.com/bomctl/bomctl/releases/download/v0.4.1/bomctl_0.4.1_darwin_amd64.tar.gz.spdx.json`
- Git Client
  - `https://git@github.com/bomctl/bomctl.git@main#sbom.cdx.json`
  - `https://git@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json`
  - `https://username:password@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json`
  - `https://git@github.com:bomctl/bomctl.git@main#sbom.cdx.json`
  - `https://github.com/bomctl/bomctl.git@main#path/to/sbom.cdx.json`
- OCI Client
  - `https://username@registry.acme.com:12345/example/image:1.2.3`
  - `https://registry.acme.com/example/image:1.2.3`

</details>

### Comparison of Example location strings

- HTTP Client
  - Current: `example.acme.com`
    - Purl: `pkg:generic/acme?download_url=example.acme.com`
    - URL: `https://example.acme.com`
  - Current: `https://github.com/bomctl/bomctl/releases/download/v0.4.1/bomctl_0.4.1_darwin_amd64.tar.gz.spdx.json`
    - Purl: `pkg:generic/bomctl@0.4.1?download_url=https://github.com/bomctl/bomctl/releases/download/v0.4.1/bomctl_0.4.1_darwin_amd64.tar.gz.spdx.json`
    - URL: `https://github.com/bomctl/bomctl/releases/download/v0.4.1/bomctl_0.4.1_darwin_amd64.tar.gz.spdx.json`
    - Notes: If we want the future github client to take this:
      - `pkg:github/bomctl/bomctl@0.4.1#bomctl_0.4.1_darwin_amd64.tar.gz.spdx.json`
- Git Client (Although these examples could use the future github client with purls also)
  - Current: `git+https://git@github.com/bomctl/bomctl.git@main#sbom.cdx.json`
    - Purl: `pkg:git/bomct/bomctl@main?user=git#sbom.cdx.json`
    - URL: `https://git@github.com/bomctl/bomctl.git@main#sbom.cdx.json`
  - Current: `ssh://git@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json`
    - Purl: `pkg:git/bomct/bomctl@main?user=git&port=12345&scheme=ssh#sbom.cdx.json`
    - URL: `https://git@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json`
  - Current: `git://username:password@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json`
    - Purl: `pkg:git/bomct/bomctl@main?user=git&password=password&port=12345&scheme=git#sbom.cdx.json`
    - URL: `https://username:password@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json`
  - Current: `git@github.com:bomctl/bomctl.git@main#sbom.cdx.json`
    - Purl: `pkg:git/bomct/bomctl@main?user=git#sbom.cdx.json`
    - URL: `https://git@github.com:bomctl/bomctl.git@main#sbom.cdx.json`
  - Current: `https://github.com/bomctl/bomctl.git@main#path/to/sbom.cdx.json`
    - Purl: `pkg:git/bomct/bomctl@main?user=git#path/to/sbom.cdx.json`
    - URL: `https://github.com/bomctl/bomctl.git@main#path/to/sbom.cdx.json`
- OCI Client
  - Current: `oci://username@registry.acme.com:12345/example/image:1.2.3`
    - Purl: `pkg:oci/example/image@1.2.3?repository_url=registry.acme.com&username=username&port=12345`
    - URL: `https://username@registry.acme.com:12345/example/image:1.2.3`
  - Current: `registry.acme.com/example/image:1.2.3`
    - Purl: `pkg:oci/example/image@1.2.3?repository_url=registry.acme.com`
    - URL: `https://registry.acme.com/example/image:1.2.3`

## Decision

## Consequences
