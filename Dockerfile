FROM golang:1.21.0-bullseye as builder
MAINTAINER kalikaneko@riseup.net

COPY . /workdir
WORKDIR /workdir

ENV CGO_CPPFLAGS="-D_FORTIFY_SOURCE=2 -fstack-protector-all"
ENV GOFLAGS="-buildmode=pie"

RUN go build -ldflags "-s -w" -trimpath .

FROM debian:12-slim

COPY --from=builder /workdir/ghosthugo /usr/local/bin/ghosthugo
COPY ghosthugo_entrypoint.sh /usr/local/bin/ghosthugo_entrypoint.sh

ENV HUGO_VERSION='0.121.2'
ENV HUGO_NAME="hugo_extended_${HUGO_VERSION}_Linux-amd64"
ENV HUGO_URL="https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/${HUGO_NAME}.deb"
ENV BUILD_DEPS="wget"
RUN mkdir /site
VOLUME /site

RUN apt-get update && \
    apt-get install -y nodejs npm && \
    apt-get install -y git "${BUILD_DEPS}" && \
    wget "${HUGO_URL}" && \
    apt-get install "./${HUGO_NAME}.deb" && \
    rm -rf "./${HUGO_NAME}.deb" "${HUGO_NAME}" && \
    apt-get install -y ssh-client rsync && \
    apt-get remove -y "${BUILD_DEPS}" && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

#USER 1000

ENTRYPOINT [ "/usr/local/bin/ghosthugo_entrypoint.sh" ]



