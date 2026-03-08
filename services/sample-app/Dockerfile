# Build from repo root: docker build -f services/sample-app/Dockerfile -t sample-app:latest .
FROM golang:1.25-alpine AS builder

WORKDIR /workspace

# Copy shared contracts module first (required by replace directive in go.mod)
COPY contracts/ contracts/

# Copy service source
COPY services/sample-app/ services/sample-app/

WORKDIR /workspace/services/sample-app
RUN go build -o /bin/sample-app .

FROM alpine:3.19
COPY --from=builder /bin/sample-app /bin/sample-app
EXPOSE 8080
ENTRYPOINT ["/bin/sample-app"]
