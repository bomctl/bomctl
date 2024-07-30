# SBOM Diff

## Context and Problem Statement

Seeing where two SBOMs differ and highlighting potentially problematic changes such as: 
- a dependency that is slightly renamed and pointing to a different url
- a new dependency that is pointing to a new/unknown package 
- a dependency whose version is bumped down

 Some tools do that do that on the package lock level since that sort of supply chain attack is starting to get a bit more prevalent.

Are there other useful things we should highlight for the user?

## Considered Options

* display various potentially problematic changes where the SBOMs differ

## Decision Outcome

Chosen option: "display various potentially problematic changes where the SBOMs differ", because the functionality could be useful and give actionable and useful information to the end user regarding SBOMs
