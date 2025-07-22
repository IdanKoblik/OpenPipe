FROM alpine

LABEL org.opencontainers.image.authors="me@idank.dev"

RUN adduser -D -u 10001 container

WORKDIR /app/

COPY openPipe /app/

USER container
WORKDIR /home/container/

ENTRYPOINT ["/app/openPipe", "--docker"]

