ARG MODE
FROM postgres:latest AS pg-tls
COPY pg_hba.conf /var/lib/postgresql/pg_hba.conf
COPY certs/rootCA.crt /var/lib/postgresql/rootCA.crt
COPY certs/postgres.server.crt /var/lib/postgresql/postgres.server.crt
COPY certs/postgres.server.key /var/lib/postgresql/postgres.server.key
RUN chown -R postgres /var/lib/postgresql && \
	chmod 600 /var/lib/postgresql/postgres.server.key
CMD ["-c", "ssl=on", "-c", "ssl_cert_file=/var/lib/postgresql/postgres.server.crt", "-c", "ssl_key_file=/var/lib/postgresql/postgres.server.key", "-c", "ssl_ca_file=/var/lib/postgresql/rootCA.crt", "-c", "hba_file=/var/lib/postgresql/pg_hba.conf"]
FROM postgres:latest AS pg-standard

FROM pg-${MODE}
