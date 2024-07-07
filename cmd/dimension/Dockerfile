FROM golang:1.19-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 1
ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn,direct
RUN apk --no-cache update && \
    apk --no-cache add tzdata && \
    apk --no-cache add git gcc libc-dev librdkafka-dev pkgconf

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -tags musl -ldflags "-s -w" -o /app/dimension cmd/dimension/dimension.go


FROM alpine:3.17

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/dimension /app/dimension
COPY --from=builder /tmp /tmp

# 暴露端口: 3000 是默认的pprof的端口
EXPOSE 3000

ENTRYPOINT ["/app/dimension"]
