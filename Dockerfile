# syntax = docker/dockerfile:latest
FROM golang as build
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go mod tidy; \
  CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o lazytainer ./...

FROM scratch
COPY --from=build /app/lazytainer /usr/local/bin/lazytainer
ENTRYPOINT ["/usr/local/bin/lazytainer"]
