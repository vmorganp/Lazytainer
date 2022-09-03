FROM golang:alpine as build
RUN apk add --no-cache g++
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -trimpath -ldflags="-s -w" -o lazytainer ./...

FROM alpine
COPY --from=build /app/lazytainer /usr/local/bin/lazytainer
ENTRYPOINT ["/usr/local/bin/lazytainer"]
