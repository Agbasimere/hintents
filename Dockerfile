# Stage 1: Build Rust simulator
FROM --platform=$BUILDPLATFORM rust:alpine AS builder-rust

ARG TARGETPLATFORM
ARG BUILDPLATFORM

WORKDIR /app/simulator

# Install build dependencies for cross-compilation
RUN apk add --no-cache musl-dev gcc

# Copy Rust project files
COPY simulator/Cargo.toml simulator/Cargo.lock ./
COPY simulator/src ./src

# Build release binary (statically linked by default on Alpine)
# Rust automatically handles the target architecture
RUN cargo build --release

# Stage 2: Build Go CLI
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder-go

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Copy Go dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy Go source
COPY . .

# Build Go binary statically for target architecture
ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
RUN go build -ldflags="-s -w" -o erst cmd/erst/main.go

# Stage 3: Final Runtime Image
FROM alpine:latest

ARG VERSION
ARG COMMIT_SHA
ARG BUILD_DATE

LABEL org.opencontainers.image.title="erst"
LABEL org.opencontainers.image.description="Execution Runtime Simulation Tool with Go CLI and Rust simulator"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${COMMIT_SHA}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.source="https://github.com/dotandev/hintents"
LABEL org.opencontainers.image.licenses="Apache-2.0"

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binaries from builders
COPY --from=builder-go /app/erst .
COPY --from=builder-rust /app/simulator/target/release/erst-sim ./simulator/target/release/erst-sim

# Verify binaries are executable
RUN chmod +x ./erst ./simulator/target/release/erst-sim

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ./erst --version || exit 1

ENTRYPOINT ["./erst"]
