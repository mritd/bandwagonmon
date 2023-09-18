FROM golang:1-alpine AS builder

RUN set -e \
    && apk upgrade \
    && apk add git

COPY . /go/src/github.com/mritd/bandwagonmon

WORKDIR /go/src/github.com/mritd/bandwagonmon

RUN set -ex \
    && apk add gcc musl-dev \
    && go install -trimpath -ldflags "-w -s"
    

FROM alpine AS dist

LABEL maintainer="mritd <mritd@linux.com>"

ENV TZ Asia/Shanghai

COPY --from=builder /go/bin/bandwagonmon /usr/local/bin/bandwagonmon

# set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/golang/go/blob/go1.9.1/src/net/conf.go#L194-L275
# - docker run --rm debian:stretch grep '^hosts:' /etc/nsswitch.conf
RUN echo 'hosts: files dns' > /etc/nsswitch.conf

RUN set -e \
    && apk upgrade \
    && apk add bash tzdata \
    && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && rm -rf /var/cache/apk/*

EXPOSE 8080

CMD ["bandwagonmon"]
