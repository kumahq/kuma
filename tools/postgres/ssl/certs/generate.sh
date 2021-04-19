#!/bin/bash

openssl genrsa -out postgres.client.key 2048
openssl req -new -sha256 -key postgres.client.key -subj "/C=US/ST=CA/O=MyOrg, Inc./CN=kuma" -out postgres.client.csr
openssl x509 -req -in postgres.client.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out postgres.client.crt -days 500 -sha256

openssl genrsa -out postgres.server.key 2048
openssl req -new -sha256 -key postgres.server.key -subj "/C=US/ST=CA/O=MyOrg, Inc./CN=kuma" -out postgres.server.csr
openssl x509 -req -in postgres.server.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out postgres.server.crt -days 500 -sha256
