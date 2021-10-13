# from alpine
# RUN apk update && apk add --no-cache docker-cli
# COPY ./get_stats.sh /get_stats.sh

# CMD [ "/bin/sh", "/get_stats.sh" ] 
# # ENTRYPOINT ["tail", "-f", "/dev/null"]

#### old sad sh version above here
from golang:alpine3.14

WORKDIR /go/src/app
COPY ./sandman.go .
RUN go build sandman.go 

CMD ["./sandman"]