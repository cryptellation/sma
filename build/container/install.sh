#!/usr/bin/env bash

# Determinate the required values

SVC_PKG="github.com/cryptellation/sma/svc"
VERSION="$(git describe --tags --always --abbrev=0 --match='v[0-9]*.[0-9]*.[0-9]*' 2> /dev/null | sed 's/^.//')"
COMMIT_HASH="$(git rev-parse --short HEAD)"

# Build the ldflags

LDFLAGS=(
  "-X '${SVC_PKG}/service_info.Version=${VERSION}'"
  "-X '${SVC_PKG}/service_info.CommitHash=${COMMIT_HASH}'"
)

# Actual Go install process

CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go install -ldflags="${LDFLAGS[*]}" ./cmd/*