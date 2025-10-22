FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY knot .

ENTRYPOINT ["./knot"]