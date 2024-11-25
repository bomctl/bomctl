<!--
This is a template for [Documenting Architecture Decisions - Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).

You can use [adr-tools](https://github.com/npryce/adr-tools) for managing the ADR files.

In each ADR file, write the following sections.
-->
# 4. Authentication types and secrets encryption

Date: 2024-07-11

## Status
<!--
A decision may be "proposed" if the project stakeholders haven't agreed with it yet, or "accepted" once it is agreed.
If a later ADR changes or reverses a decision, it may be marked as "deprecated" or "superseded" with a reference to
its replacement.
-->
Proposed

## Context
<!--
This section describes the forces at play, including technological, political, social, and project local. These forces
are probably in tension, and should be called out as such. The language in this section is value-neutral. It is simply
describing facts.
-->
Additional authentication mechanisms and secure secrets storage should be supported.

## Decision
<!--
This section describes our response to these forces. It is stated in full sentences, with active voice. "We will …"
-->
Add support for additional authentication mechanisms:

- OAuth
- Bearer token
- Basic

Secrets provided as plain text will be encrypted with a user-provided or auto-generated key pair.

Add a config file mapping of URLs or bare hostnames to associated user credentials.

- If credentials are specified by the user or encountered in the config file as plain text:
  - If no private key is provided by user or already exists in the config directory,
    auto-generate a new key pair
  - Encrypt secrets inline in the config file using the private key

Either leverage the [SOPS](https://getsops.io) tool as a library to perform the encryption/decryption,
or use its encrypted string expression form.

Example of proposed config file additions:

```yaml
auths:
  github.com: ENC[AES256_GCM,data:Tr7o=,iv:1=,aad:No=,tag:k=]
  gitlab.com:
    user: ENC[AES256_GCM,data:CwE4O1s=,iv:2k=,aad:o=,tag:w==]
    password: ENC[AES256_GCM,data:p673w==,iv:YY=,aad:UQ=,tag:A=]
```

## Consequences
<!--
This section describes the resulting context, after applying the decision. All consequences should be listed here, not
just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them
affect the team and project in the future.
-->
These changes will increase flexibility for users by allowing fetching from and pushing to additional
remote endpoints that may have limited or strict options for authentication. They will also provide
enhanced security options for local storage and transmission of secrets.
