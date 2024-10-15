# Build binary
FROM --platform=$BUILDPLATFORM golang:1.17.1-alpine AS build-env
ADD . /app
WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o microcks-cli github.com/microcks/microcks-cli
    
# Build image
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.4-1227.1726694542

# Some version information
LABEL maintainer="Laurent Broudoux <laurent@microcks.io>" \
      org.opencontainers.image.authors="Laurent Broudoux <laurent@microcks.io>" \
      org.opencontainers.image.title="Microcks CLI" \
      org.opencontainers.image.description="Microcks is Open Source cloud-native native tool for API Mocking and Testing" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.documentation="https://github.com/microcks/microcks-cli" \
      io.artifacthub.package.readme-url="https://raw.githubusercontent.com/microcks/microcks-cli/master/README.md"

ENV CLI=/usr/local/bin/microcks-cli \
    USER_UID=1001 \
    USER_NAME=microcks-cli

# install cli binary
COPY --from=build-env /app/microcks-cli ${CLI}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

USER ${USER_UID}