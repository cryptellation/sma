# Dockerfile arguments
ARG BUILDPLATFORM=linux/amd64

ARG TARGETOS=linux
ARG TARGETARCH=amd64

ARG BUILDBASEIMAGE=golang:alpine
ARG TARGETBASEIMAGE=alpine:latest

# Building Golang image
FROM --platform=${BUILDPLATFORM:-linux/amd64} ${BUILDBASEIMAGE} AS build

# Disable CGO
ENV CGO_ENABLED=0

# Install requirements
RUN apk add --no-cache git bash

# Get all remaining code
RUN mkdir -p /go/src/github.com/cryptellation/sma
COPY ./ /go/src/github.com/cryptellation/sma

# Set the workdir
WORKDIR /go/src/github.com/cryptellation/sma

# Build everything in cmd/
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    bash build/container/install.sh

# Get final base image
FROM --platform=${TARGETOS}/${TARGETARCH} ${TARGETBASEIMAGE} AS final

# Set environment variables
ENV HEALTH_PORT=9000

# Expose ports
EXPOSE 9000

# Get binary
COPY --from=build /go/bin/* /usr/local/bin

# Set command
CMD ["worker", "serve"]