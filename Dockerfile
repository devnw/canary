# Multi-stage build for llm-ssh server at repo root
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install required tools for fetching modules
RUN apk add --no-cache git

COPY . .

# Build static-ish binary
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o /server .

# Runtime image
FROM alpine:latest

# CA certs for HTTPS to OpenAI and auth endpoint
RUN apk add --no-cache ca-certificates && update-ca-certificates

# App
COPY --from=builder /server /server

# Config (no secrets baked in)
COPY --from=builder /app/scenarios.yaml /scenarios.yaml
COPY --from=builder /app/emulator.md /emulator.md

# Defaults (can override at runtime)
ENV PORT=2029 \
    HOST=0.0.0.0 \
    SCENARIOS_FILE=/scenarios.yaml \
    SCENARIO_INDEX=1 \
    OPENAI_MODEL=gpt-4o-mini

EXPOSE 2029

CMD ["/server"]
