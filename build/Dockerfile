# Build binary
FROM --platform=$BUILDPLATFORM golang:1.17.1-alpine AS build-env
ADD . /app
WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o microcks-cli github.com/microcks/microcks-cli
    
# Build image
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.2-750.1696515534

MAINTAINER Laurent Broudoux <laurent@microcks.io>

ENV CLI=/usr/local/bin/microcks-cli \
    USER_UID=1001 \
    USER_NAME=microcks-cli

# install cli binary
COPY --from=build-env /app/microcks-cli ${CLI}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

USER ${USER_UID}