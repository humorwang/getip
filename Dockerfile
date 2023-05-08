ARG GO_VERSION=1.18

FROM golang:${GO_VERSION}-alpine AS builder

RUN mkdir -p /app
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct

COPY . .
RUN go build -o ./app main.go


FROM alpine:3.14

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

RUN mkdir -p /app/geolite2
WORKDIR /app
COPY --from=builder /app/app ./
COPY --from=builder /app/geolite2/ ./geolite2/

EXPOSE 8080
ENTRYPOINT ["./app"]