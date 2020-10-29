ARG BUILD_FROM

FROM golang:1.15-alpine3.12 AS builder

WORKDIR /usr/src
ARG BUILD_ARCH
ARG COREDNS_VERSION

# Build CoreDNS
COPY plugins plugins
RUN \
    set -x \
    && apk add --no-cache \
        git \
        make \
        bash \
    && git clone --depth 1 -b v${COREDNS_VERSION} https://github.com/coredns/coredns \
    && cp -rf plugins/* coredns/plugin/ \
    && cd coredns \
    && sed -i "/^forward:.*/a fallback:fallback" plugin.cfg \
    && sed -i "/^hosts:.*/a mdns:mdns" plugin.cfg \
    && go generate \
    && \
        if [ "${BUILD_ARCH}" = "armhf" ]; then \
            make coredns SYSTEM="CGO_ENABLED=0 GOOS=linux GOARM=6 GOARCH=arm"; \
        elif [ "${BUILD_ARCH}" = "armv7" ]; then \
            make coredns SYSTEM="CGO_ENABLED=0 GOOS=linux GOARM=7 GOARCH=arm"; \
        elif [ "${BUILD_ARCH}" = "aarch64" ]; then \
            make coredns SYSTEM="CGO_ENABLED=0 GOOS=linux GOARCH=arm64"; \
        elif [ "${BUILD_ARCH}" = "i386" ]; then \
            make coredns SYSTEM="CGO_ENABLED=0 GOOS=linux GOARCH=386"; \
        elif [ "${BUILD_ARCH}" = "amd64" ]; then \
            make coredns SYSTEM="CGO_ENABLED=0 GOOS=linux GOARCH=amd64"; \
        else \
            exit 1; \
        fi

FROM ${BUILD_FROM}

WORKDIR /config
COPY --from=builder /usr/src/coredns/coredns /usr/bin/coredns
COPY rootfs /
