# syntax = docker/dockerfile:latest
FROM golang:1.22.3-alpine3.19 as build
RUN apk add --update build-base gcc wget git libpcap-dev
WORKDIR /app
COPY src/* /app/
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go mod tidy; \
  CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o lazytainer ./...

FROM alpine
RUN apk add --update libpcap-dev
COPY --from=build /app/lazytainer /app/lazytainer
ENTRYPOINT [ "./app/lazytainer" ]

