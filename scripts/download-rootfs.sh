#!/bin/bash
set -euo pipefail

ALPINE_VERSION="3.21"
ALPINE_MINOR="3.21.3"
ARCH="x86_64"
DEST="/var/lib/minic/images/alpine/rootfs"

URL="https://dl-cdn.alpinelinux.org/alpine/v${ALPINE_VERSION}/releases/${ARCH}/alpine-minirootfs-${ALPINE_MINOR}-${ARCH}.tar.gz"

echo "Downloading Alpine ${ALPINE_MINOR} minirootfs..."
mkdir -p "${DEST}"
curl -fSL "${URL}" | tar -xz -C "${DEST}"
echo "Rootfs extracted to ${DEST}"
