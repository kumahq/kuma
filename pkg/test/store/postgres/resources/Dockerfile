FROM postgres:latest
COPY pg_hba.conf /var/lib/postgresql/pg_hba.conf
COPY certs/rootCA.crt /var/lib/postgresql/rootCA.crt
COPY certs/postgres.server.crt /var/lib/postgresql/postgres.server.crt
COPY certs/postgres.server.key /var/lib/postgresql/postgres.server.key
RUN chown postgres /var/lib/postgresql/postgres.server.key && \
    chmod 600 /var/lib/postgresql/postgres.server.key
