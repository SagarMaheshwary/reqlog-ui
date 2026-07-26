FROM --platform=$BUILDPLATFORM golang:1.25 AS reqlog-builder

ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG REQLOG_VERSION

RUN CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go install github.com/sagarmaheshwary/reqlog/cmd/reqlog@${REQLOG_VERSION}

FROM debian:bookworm-slim AS production

ARG VERSION

LABEL org.opencontainers.image.title="reqlog-ui"
LABEL org.opencontainers.image.description="A lightweight web UI for reqlog — search and trace logs directly from your browser."
LABEL org.opencontainers.image.source="https://github.com/sagarmaheshwary/reqlog-ui"
LABEL org.opencontainers.image.url="https://github.com/sagarmaheshwary/reqlog-ui"
LABEL org.opencontainers.image.documentation="https://github.com/sagarmaheshwary/reqlog-ui/blob/main/README.md"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.version="${VERSION}"

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        gnupg && \
    install -m 0755 -d /etc/apt/keyrings && \
    curl -fsSL https://download.docker.com/linux/debian/gpg \
        | gpg --dearmor -o /etc/apt/keyrings/docker.gpg && \
    chmod a+r /etc/apt/keyrings/docker.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian bookworm stable" \
        > /etc/apt/sources.list.d/docker.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends docker-ce-cli && \
    apt-get purge -y curl gnupg && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=reqlog-builder /go/bin/reqlog /usr/local/bin/reqlog
COPY reqlog-ui ./main

EXPOSE 4000

CMD ["./main"]