FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o aperture ./cmd/server

# Runtime is Ubuntu (glibc 2.39) rather than Alpine because optional AMD GPU
# monitoring relies on amd-smi — a python3 script backed by libamd_smi.so, which
# needs glibc >= 2.38 and cannot run on musl. The Go binary is statically linked
# (CGO_ENABLED=0), so it runs here unchanged.
FROM ubuntu:24.04
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates wget python3 libstdc++6 libdrm2 \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/aperture /app/aperture
# Optional AMD GPU support: when the host passes through /dev/kfd + /dev/dri and
# mounts its ROCm tree read-only at /opt/rocm, amd-smi is found here on PATH and
# the GPU adapter reports utilization / VRAM / temperature. Without the mount the
# adapter simply probes, finds nothing, and reports "Not available".
ENV PATH="/opt/rocm/bin:${PATH}"
EXPOSE 8888
ENTRYPOINT ["/app/aperture"]
