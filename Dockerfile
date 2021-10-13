from alpine
RUN apk update && apk add --no-cache docker-cli
COPY ./get_stats.sh /get_stats.sh

CMD [ "/bin/sh", "/get_stats.sh" ] 
# ENTRYPOINT ["tail", "-f", "/dev/null"]