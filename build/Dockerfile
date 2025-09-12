# Build binary
FROM --platform=$BUILDPLATFORM golang:1.23.0-alpine AS build-env
ADD . /app
WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o microcks github.com/microcks/microcks-cli
    
# Build image
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.6-1755695350

# Some version information
LABEL maintainer="Laurent Broudoux <laurent@microcks.io>" \
      org.opencontainers.image.authors="Laurent Broudoux <laurent@microcks.io>" \
      org.opencontainers.image.title="Microcks CLI" \
      org.opencontainers.image.description="Microcks is Open Source cloud-native native tool for API Mocking and Testing" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.documentation="https://github.com/microcks/microcks-cli" \
      io.artifacthub.package.readme-url="https://raw.githubusercontent.com/microcks/microcks-cli/master/README.md"

ENV CLI=/usr/local/bin/microcks \
    USER_UID=1001 \
    USER_NAME=microcks-cli

# install cli binary
COPY --from=build-env /app/microcks ${CLI}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

USER ${USER_UID}