FROM ghcr.io/tinfoilsh/nitro-attestation-shim:v0.2.2 AS shim

FROM golang:1.22 AS build
WORKDIR /app
COPY main.go go.mod ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /contentmod main.go

FROM ollama/ollama

RUN apt-get update && apt-get install -y iproute2

# Copy in the shim, your Go binary, and the start script
COPY --from=shim   /nitro-attestation-shim /nitro-attestation-shim
COPY --from=build  /contentmod            /contentmod
COPY start.sh /start.sh
RUN chmod +x /start.sh

ENV HOME=/
ENV PORT=80

# Wrap your start.sh with the shim
ENTRYPOINT ["/nitro-attestation-shim", "-e", "tls@tinfoil.sh", "-u", "80", "--", "/start.sh"]
