ARG MODE
# With test-containers building the same image many times have race conditions when deleting the images
# We therefore add a unique ID just to make images different
ARG UNIQUEID
FROM postgres:latest@sha256:edb0c08dc33befd0262ab3b5cf1c4f3f8cfd8b09e88c05b45adee29d6761ddb6 AS pg-tls
COPY pg_hba.conf /etc/postgresql/pg_hba.conf
COPY certs/rootCA.crt /etc/postgresql/rootCA.crt
COPY certs/postgres.server.crt /etc/postgresql/postgres.server.crt
COPY certs/postgres.server.key /etc/postgresql/postgres.server.key
RUN chown -R postgres:postgres /etc/postgresql && \
	chmod 600 /etc/postgresql/postgres.server.key
CMD ["-c", "ssl=on", "-c", "max_connections=10000", "-c", "ssl_cert_file=/etc/postgresql/postgres.server.crt", "-c", "ssl_key_file=/etc/postgresql/postgres.server.key", "-c", "ssl_ca_file=/etc/postgresql/rootCA.crt", "-c", "hba_file=/etc/postgresql/pg_hba.conf"]
FROM postgres:latest@sha256:edb0c08dc33befd0262ab3b5cf1c4f3f8cfd8b09e88c05b45adee29d6761ddb6 AS pg-standard
CMD ["-c", "max_connections=10000"]

FROM pg-${MODE}
RUN echo ${UNIQUEID} > /tmp/uniqueID
