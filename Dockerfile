FROM ghcr.io/tinfoilanalytics/nitro-attestation-shim:v0.2.2 AS shim

FROM ollama/ollama AS ollama

FROM golang:1.22 AS build
WORKDIR /app
COPY main.go go.mod ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /contentmod main.go

FROM alpine:3
RUN apk add --no-cache iproute2 ca-certificates

# Copy in the shim, the Ollama binary, and your Go app
COPY --from=shim   /nitro-attestation-shim /nitro-attestation-shim
COPY --from=ollama /bin/ollama            /bin/ollama
COPY --from=build  /contentmod            /contentmod

# Copy the start script
COPY start.sh /start.sh
RUN chmod +x /start.sh

ENV HOME=/
ENV PORT=80

ENTRYPOINT ["/nitro-attestation-shim", "-e", "tls@tinfoil.sh", "-u", "80", "--", "/start.sh"]