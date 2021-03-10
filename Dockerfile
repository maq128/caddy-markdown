FROM golang:1.15.8-alpine as builder

RUN echo http://mirrors.aliyun.com/alpine/latest-stable/main/ > /etc/apk/repositories \
 && echo http://mirrors.aliyun.com/alpine/latest-stable/community/ >> /etc/apk/repositories \
 && apk add git build-base \
 && go get github.com/caddyserver/xcaddy/cmd/xcaddy \
 && xcaddy build v2.3.0 --with github.com/maq128/caddy-markdown@latest

FROM alpine

RUN echo http://mirrors.aliyun.com/alpine/latest-stable/main/ > /etc/apk/repositories \
 && echo http://mirrors.aliyun.com/alpine/latest-stable/community/ >> /etc/apk/repositories \
 && apk update \
 && apk add tzdata ca-certificates \
 && rm -rf /var/cache/apk/*

COPY --from=builder /go/caddy /usr/local/bin/
ENV XDG_CONFIG_HOME /data
ENV XDG_DATA_HOME /data
VOLUME /data

EXPOSE 80
EXPOSE 443

WORKDIR /srv

CMD ["caddy", "run"]
