# Build from repo root: docker build -f services/batch-worker/Dockerfile -t batch-worker:latest .
FROM golang:1.25-alpine AS builder

WORKDIR /workspace

# Copy shared contracts module first (required by replace directive in go.mod)
COPY contracts/ contracts/

# Copy service source
COPY services/batch-worker/ services/batch-worker/

WORKDIR /workspace/services/batch-worker
RUN go build -o /bin/batch-worker .

FROM alpine:3.19
COPY --from=builder /bin/batch-worker /bin/batch-worker
ENTRYPOINT ["/bin/batch-worker"]
