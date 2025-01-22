FROM ghcr.io/tinfoilanalytics/nitro-attestation-shim:v0.2.2 AS shim

FROM ollama/ollama AS ollama

FROM golang:1.21 AS build

WORKDIR /app
COPY main.go go.mod go.sum ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /contentmod main.go

FROM alpine:3

RUN apk add --no-cache iproute2 ca-certificates

# Copy binaries from previous stages
COPY --from=shim /nitro-attestation-shim /nitro-attestation-shim
COPY --from=ollama /bin/ollama /bin/ollama
COPY --from=build /contentmod /contentmod

# Set environment variable for port
ENV PORT=80

# Create script to initialize ollama and start the application
RUN echo '#!/bin/sh\n\
    ollama serve &\n\
    sleep 5\n\
    ollama pull llama-guard3:1b\n\
    exec "$@"' > /start.sh && chmod +x /start.sh

ENTRYPOINT ["/start.sh", "/nitro-attestation-shim", "-e", "tls@tinfoil.sh", "-u", "80", "--", "/contentmod"]