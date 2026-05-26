FROM golang:1.26-alpine AS server

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

RUN go.mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -ldflags="-s -w" \
    -o ./server ./cmd/orchestrator


FROM alpine:3.21 AS cloudflare-tunnel
RUN apk add --no-cache curl

RUN curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64 \
    -o /usr/local/bin/cloudflared && \
    chmod +x /usr/local/bin/cloudflared

FROM alpine:3.21

WORKDIR /app

COPY --from=server /app/server ./server
COPY --from=cloudflare-tunnel /usr/local/bin/cloudflared /usr/local/bin/cloudflared

RUN echo '#!/bin/sh \n\
./server & \n\
exec /usr/local/bin/cloudflared tunnel \n\
  --protocol quic \n\
  --edge-ip-version 6 \n\
  --post-quantum \n\
  run --token "$TUNNEL_TOKEN" \n\
' > ./run.sh && chmod +x ./run.sh

# Document the backend application's target listening port
EXPOSE 8080

ENTRYPOINT ["./run.sh"]