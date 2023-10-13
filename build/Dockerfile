FROM registry.access.redhat.com/ubi9/ubi-minimal:9.2-750.1696515534

MAINTAINER Laurent Broudoux <laurent@microcks.io>

ENV CLI=/usr/local/bin/microcks-cli \
    USER_UID=1001 \
    USER_NAME=microcks-cli

# install cli binary
COPY /build/_output/microcks-cli-linux-amd64 ${CLI}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

USER ${USER_UID}