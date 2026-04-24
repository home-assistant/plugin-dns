FROM golang:1.25.7-alpine3.23 AS builder

WORKDIR /usr/src
ARG TARGETARCH
ARG COREDNS_VERSION="1.11.4"

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
    && sed -i "/^template:template/d" plugin.cfg \
    && sed -i "/^hosts:.*/a template:template" plugin.cfg \
    && sed -i "/^forward:.*/i fallback:fallback" plugin.cfg \
    && sed -i "/^hosts:.*/a mdns:mdns" plugin.cfg \
    && sed -i "/route53:route53/d" plugin.cfg \
    && sed -i "/clouddns:clouddns/d" plugin.cfg \
    && sed -i "/k8s_external:k8s_external/d" plugin.cfg \
    && sed -i "/kubernetes:kubernetes/d" plugin.cfg \
    && sed -i "/etcd:etcd/d" plugin.cfg \
    && sed -i "/grpc:grpc/d" plugin.cfg \
    && go mod tidy \
    && go generate \
    && if [ -z "${TARGETARCH}" ]; then \
            echo "TARGETARCH is not set, please use Docker BuildKit for the build." && exit 1; \
        fi \
    && case "${TARGETARCH}" in \
            amd64|arm64) ;; \
            *) echo "Unsupported TARGETARCH: ${TARGETARCH}" && exit 1 ;; \
        esac \
    && make coredns SYSTEM="CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH}"

# Base image updated by Renovate, update versionCompatibility on Alpine base bump
FROM ghcr.io/home-assistant/base:3.23-2026.04.0@sha256:5ea71eca9735949b080a05d271988c9663af3bd1050297abedefa222c61e6807

WORKDIR /config
COPY --from=builder /usr/src/coredns/coredns /usr/bin/coredns
COPY rootfs /

LABEL \
    io.hass.type="dns" \
    org.opencontainers.image.title="Home Assistant DNS Plugin" \
    org.opencontainers.image.description="Home Assistant Supervisor plugin for DNS" \
    org.opencontainers.image.authors="The Home Assistant Authors" \
    org.opencontainers.image.url="https://www.home-assistant.io/" \
    org.opencontainers.image.documentation="https://www.home-assistant.io/docs/" \
    org.opencontainers.image.licenses="Apache License 2.0"
