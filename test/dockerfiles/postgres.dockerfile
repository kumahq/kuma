ARG MODE
# With test-containers building the same image many times have race conditions when deleting the images
# We therefore add a unique ID just to make images different
ARG UNIQUEID
FROM postgres:latest@sha256:a9abf4275f9e99bff8e6aed712b3b7dfec9cac1341bba01c1ffdfce9ff9fc34a AS pg-tls
COPY pg_hba.conf /etc/postgresql/pg_hba.conf
COPY certs/rootCA.crt /etc/postgresql/rootCA.crt
COPY certs/postgres.server.crt /etc/postgresql/postgres.server.crt
COPY certs/postgres.server.key /etc/postgresql/postgres.server.key
RUN chown -R postgres:postgres /etc/postgresql && \
	chmod 600 /etc/postgresql/postgres.server.key
CMD ["-c", "ssl=on", "-c", "max_connections=10000", "-c", "ssl_cert_file=/etc/postgresql/postgres.server.crt", "-c", "ssl_key_file=/etc/postgresql/postgres.server.key", "-c", "ssl_ca_file=/etc/postgresql/rootCA.crt", "-c", "hba_file=/etc/postgresql/pg_hba.conf"]
FROM postgres:latest@sha256:a9abf4275f9e99bff8e6aed712b3b7dfec9cac1341bba01c1ffdfce9ff9fc34a AS pg-standard
CMD ["-c", "max_connections=10000"]

FROM pg-${MODE}
RUN echo ${UNIQUEID} > /tmp/uniqueID
