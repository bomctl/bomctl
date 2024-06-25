# Architecture Decision Records

## Create a new ADR

For new ADRs, the status should be `Proposed`. Use one of the following methods to automatically generate a markdown document for the ADR.

### `adr-tools`

1. Install [`adr-tools`](https://github.com/npryce/adr-tools/blob/master/INSTALL.md)
1. Run `adr new <title of new ADR as it should appear in documentation>`
1. Open the new file that was generated in this directory
1. Update the placeholder text in the new file to provide information for all required fields

### VSCode Snippet

1. Create a new markdown file in this directory. The file should follow the naming format `NNNN-lowercase-title-with-dashes.md`
1. Open the file in a VSCode editor
1. Type `adr` and select it from the editor suggestions (may need to show suggestions with `CTRL + space` or `command + space`)
1. Use tab to navigate through the fields requiring input and update the placeholder text

## Submit ADR

Follow the steps outlined in [CONTRIBUTING.md](../../CONTRIBUTING.md#steps-to-contribute) to create a pull request that adds the new ADR file to the repository.
