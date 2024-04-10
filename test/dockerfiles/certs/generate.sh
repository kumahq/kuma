#!/usr/bin/env bash
openssl genrsa -out rootCA.key 4096
openssl req -x509 -new -nodes -key rootCA.key -sha256 -days 10024 -out rootCA.crt -subj "/C=US/ST=CA/O=MyOrg, Inc./CN=kuma-ca"

openssl genrsa -out postgres.client.key 2048
openssl req -new -sha256 -key postgres.client.key -subj "/C=US/ST=CA/O=MyOrg, Inc./CN=kuma" -out postgres.client.csr
openssl x509 -req -in postgres.client.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out postgres.client.crt -days 5000 -sha256

openssl genrsa -out postgres.server.key 2048
openssl req -new -sha256 -key postgres.server.key -subj "/C=US/ST=CA/O=MyOrg, Inc./CN=kuma-server" -out postgres.server.csr
openssl x509 -req -in postgres.server.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -extfile <(printf "subjectAltName = DNS:localhost") -out postgres.server.crt -days 5000 -sha256
