from golang:alpine3.14
WORKDIR /go/src/app
COPY ./sandman.go .
RUN go build sandman.go

FROM alpine:latest
RUN apk --no-cache add docker-cli
WORKDIR /root/
COPY --from=0 /go/src/app/sandman ./
ENV TIMEOUT=30
ENV LABEL=sandman


CMD ["./sandman"]