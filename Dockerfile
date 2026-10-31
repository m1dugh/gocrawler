# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3

# The build stage always runs on the machine doing the build, whatever the
# target platform is, and cross-compiles: the binary is pure Go, so building
# for amd64 or arm64 needs no emulation here.
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS build

ENV CGO_ENABLED=0 \
    GOTOOLCHAIN=local

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# Set by BuildKit from --platform. They are empty with the legacy builder,
# which then builds for the host.
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" \
        -o /out/gocrawler ./cmd/gocrawler

# The final stage is built for each target platform.
FROM alpine:${ALPINE_VERSION}

# No browser is installed on purpose, to keep the image small: from this image
# the browser engine is only used with an existing browser, through
# --browser-url. This is an architectural choice that might change.
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 gocrawler

COPY --from=build /out/gocrawler /usr/local/bin/gocrawler

USER gocrawler
WORKDIR /work

ENTRYPOINT ["gocrawler"]
