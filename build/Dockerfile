FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

MAINTAINER Laurent Broudoux <laurent.broudoux@gmail.com>

ENV CLI=/usr/local/bin/microcks-cli \
    USER_UID=1001 \
    USER_NAME=microcks-cli

# install cli binary
COPY /build/_output/microcks-cli-linux-amd64 ${CLI}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

USER ${USER_UID}