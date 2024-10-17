#!/bin/sh
set -e

redis-cli --tls \
 --cert /tls/ssl/server-cert.pem \
 --cacert /tls/ssl/ca-cert.pem \
 --key /tls/ssl/server-key.pem \
 -a '611fbff225e20a58d45641394482f26a' \
 -p 6379 \
 --raw incr