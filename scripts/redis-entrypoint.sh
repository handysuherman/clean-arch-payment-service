#!/bin/sh
set -e

redis-server --tls-port 6379 --port 0 \
    --tls-cert-file /tls/ssl/server-cert.pem \
    --tls-key-file /tls/ssl/server-key.pem \
    --tls-ca-cert-file /tls/ssl/ca-cert.pem \
    --tls-auth-clients yes \
    --tls-session-caching no \
    --tls-protocols "TLSv1.2 TLSv1.3" \
    --loglevel notice \
    --slowlog-max-len 128 \
    --latency-monitor-threshold 0 \
    --notify-keyspace-events "" \
    --list-max-ziplist-size -2 \
    --activerehashing yes \
    --appendfsync everysec \
    --requirepass c80990ae94e5e87d9d1061ecdf6c7650
