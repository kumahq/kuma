Generated with `kumactl generate tls-certificate --hostname=test-server.mesh --type=server`

Client certificates:
1. `openssl genpkey -algorithm RSA -out client.key -pkeyopt rsa_keygen_bits:2048`
2. `openssl req -new -key client.key -out client.csr`
3. You can use server certificate as CA `openssl x509 -req -in client.csr -CA server.crt -CAkey server.key -CAcreateserial -out client.crt -days 3650 -sha256`
