from golang:alpine3.14
WORKDIR /root/
COPY ./lazytainer.go .
RUN go build lazytainer.go

FROM alpine:latest
RUN apk --no-cache add docker-cli
WORKDIR /root/
COPY --from=0 /root/lazytainer ./

CMD ["./lazytainer"]