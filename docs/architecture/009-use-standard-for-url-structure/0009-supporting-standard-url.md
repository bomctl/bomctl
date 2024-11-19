# Investigation of using URI Standard as Command 'url'/location standard

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
