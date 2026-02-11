ARG BUILD_FROM

FROM golang:1.25.7-alpine3.23 AS builder

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
    && \
        if [ "${BUILD_ARCH}" = "aarch64" ]; then \
            make coredns SYSTEM="CGO_ENABLED=0 GOOS=linux GOARCH=arm64"; \
        elif [ "${BUILD_ARCH}" = "amd64" ]; then \
            make coredns SYSTEM="CGO_ENABLED=0 GOOS=linux GOARCH=amd64"; \
        else \
            exit 1; \
        fi

FROM ${BUILD_FROM}

WORKDIR /config
COPY --from=builder /usr/src/coredns/coredns /usr/bin/coredns
COPY rootfs /
