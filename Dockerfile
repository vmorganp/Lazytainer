FROM golang as build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -trimpath -ldflags="-s -w" lazytainer.go

FROM scratch
COPY --from=build /app/lazytainer /usr/local/bin/lazytainer
ENTRYPOINT ["/usr/local/bin/lazytainer"]
