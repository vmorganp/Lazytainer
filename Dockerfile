FROM golang:alpine3.15

ARG GOOS="linux"
ENV GOOS=${GOOS}

ARG GOARCH=""
ENV GOARCH=${GOARCH}

ENV CGO_ENABLED=0

WORKDIR /root/
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./lazytainer.go .
RUN go build lazytainer.go

FROM alpine:latest
WORKDIR /root/
COPY --from=0 /root/lazytainer ./

CMD ["./lazytainer"]