ARG MODE
# With test-containers building the same image many times have race conditions when deleting the images
# We therefore add a unique ID just to make images different
ARG UNIQUEID
FROM postgres:latest@sha256:994cc3113ce004ae73df11f0dbc5088cbe6bb0da1691dd7e6f55474202a4f211 AS pg-tls
COPY pg_hba.conf /var/lib/postgresql/pg_hba.conf
COPY certs/rootCA.crt /var/lib/postgresql/rootCA.crt
COPY certs/postgres.server.crt /var/lib/postgresql/postgres.server.crt
COPY certs/postgres.server.key /var/lib/postgresql/postgres.server.key
RUN chown -R postgres /var/lib/postgresql && \
	chmod 600 /var/lib/postgresql/postgres.server.key
CMD ["-c", "ssl=on", "-c", "max_connections=10000", "-c", "ssl_cert_file=/var/lib/postgresql/postgres.server.crt", "-c", "ssl_key_file=/var/lib/postgresql/postgres.server.key", "-c", "ssl_ca_file=/var/lib/postgresql/rootCA.crt", "-c", "hba_file=/var/lib/postgresql/pg_hba.conf"]
FROM postgres:latest@sha256:994cc3113ce004ae73df11f0dbc5088cbe6bb0da1691dd7e6f55474202a4f211 AS pg-standard
CMD ["-c", "max_connections=10000"]

FROM pg-${MODE}
RUN echo ${UNIQUEID} /tmp/uniqueID
