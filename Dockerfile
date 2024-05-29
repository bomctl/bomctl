# -------------------------------------------------------
# SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
# SPDX-FileName: Dockerfile
# SPDX-FileType: SOURCE
# SPDX-License-Identifier: Apache-2.0
# -------------------------------------------------------
FROM cgr.dev/chainguard/static:latest

ARG NONROOT_UID=65532

# 65532 is the UID of the `nonroot` user in chainguard/static.
# https://edu.chainguard.dev/chainguard/chainguard-images/reference/static/overview/#users
USER ${NONROOT_UID}:${NONROOT_UID}

COPY --chown=${NONROOT_UID}:${NONROOT_UID} bomctl /bomctl

ENTRYPOINT ["/bomctl"]
