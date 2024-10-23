# Contributing to bomctl

We're excited you're interested in contributing to the bomctl project! This document outlines the guidelines for contributing code, documentation, and other improvements to the project.

The bomctl project is under the Open Source Security Foundation (OpenSSF) Security Tooling Working Group, a collaborative effort to improve the security of open source software. We value contributions from everyone and strive to create a welcoming and inclusive community.

## Getting Started

Before diving in, here are a few things to keep in mind:

### License

The bomctl project is licensed under the [Apache 2.0 license](LICENSE). By contributing, you agree to abide by the terms of this license. You can find the license file in the repository root directory.

### Code of Conduct

We have a Code of Conduct that outlines the expectations for respectful and professional behavior in our community. Please review the [Code of Conduct](CODE_OF_CONDUCT.md) before contributing

## Ways to Contribute

There are many ways to contribute to the bomctl project:

* __Code__: Submit pull requests (PRs) for bug fixes, new features, or improvements to existing code.
* __Documentation__: Help improve the project's documentation by fixing typos, clarifying concepts, or adding new content.
* __Testing__: Write unit tests or integration tests to improve the project's code coverage and stability.
* __Reporting Issues__: If you find a bug or have a suggestion for improvement, report it as an issue on the project's GitHub repository.

## Steps to Contribute

Here's a step-by-step guide for making a contribution:

1. __Identify Issue__: Find an existing issue you want to work or submit a new issue describing your proposed change.
1. __Claim Issue__: Assign yourself to the issue if possible, or leave a comment on the issue stating your intent to work it.
1. __Fork the Repository__: Fork the bomctl repository on GitHub to your own account. This allows you to make changes to the codebase without affecting the original project.
1. __Clone the Fork__: Clone your forked repository to your local machine.
1. __Create a Branch__: Create a new branch for your changes. Use a descriptive branch name that reflects the nature of your contribution.
1. __Make Changes__: Make your changes to the codebase and write unit tests for any new features you introduce.
1. __Validate Changes__: Run the following commands or [pre-commit](https://pre-commit.com/) to to ensure the bomctl project standards are met.
    * golangci-lint (`make lint` or `make lint-fix`)
    * `go mod tidy`
    * `go test ./...`
    * `go generate ./...` and ensure any modified files are committed
1. __Commit Changes__: Commit your changes with clear and concise commit messages following the [conventional commit format](https://www.conventionalcommits.org/).
    * Your [commits must be signed](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits) with a key associated with your GitHub account.
1. __Push Changes__: Push your changes to your forked repository on GitHub.
1. __Open a Pull Request__: Open a pull request from your branch to the main branch of the upstream repository.
    * Your PR title should follow the [conventional commit format](https://www.conventionalcommits.org/).
1. __Address Reviews__: Respond to any feedback or requests for changes from the project maintainers.

## Tips for Writing Pull Requests

* Keep your pull requests focused on a single issue or feature.
* Ensure your code adheres to the project's coding style guidelines.
  * Use [pre-commit](https://pre-commit.com/) to ensure the bomctl project standards are met.
* Write clear and concise commit messages following the [convention commit format](https://www.conventionalcommits.org/) that describe the changes you made.
* Be patient and responsive to feedback from the project maintainers.

## Additional Resources

Here are some additional resources that you may find helpful:

* __GitHub Pull Requests__: <https://docs.github.com/en/pull-requests>
* __Git Basics__: <https://git-scm.com/>
* __Contributing to Open Source__: <https://www.freecodecamp.org/news/how-to-make-your-first-open-source-contribution/>

We appreciate your contributions to the bomctl project! If you have any questions, feel free to reach out to the project maintainers or open an issue on the project's GitHub repository.
