# Builder Stage
FROM golang:1.22-alpine3.18 AS builder
RUN echo "https://ftp.udx.icscoe.jp/Linux/alpine/edge/main" > /etc/apk/repositories
WORKDIR /app
COPY . .

RUN apk update
RUN go build -o main ./cmd/main.go

# Final Stage
FROM alpine:latest
RUN echo "https://ftp.udx.icscoe.jp/Linux/alpine/edge/main" > /etc/apk/repositories
RUN apk update && apk upgrade --available
WORKDIR /app
COPY --from=builder /app/main .
COPY ./configs/migration ./migration
COPY ./scripts/start.sh .
COPY ./scripts/wait-for.sh .

EXPOSE 6880
EXPOSE 7880
EXPOSE 8880
EXPOSE 50050

CMD ["/app/main", "--config-file=/app/config-cli.yaml", "--etcd-timeout=15s", "--env=production","--hostname=192.168.18.160","--uid=3f55729eac14a55c9051ef7a","--published-port=50050","--health-check-addr=192.168.18.160","--health-check-published-port=6880"]
ENTRYPOINT [ "/app/start.sh" ]