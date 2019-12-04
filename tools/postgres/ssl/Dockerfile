FROM postgres:latest
COPY certs/postgres.server.key /var/lib/postgresql/postgres.server.key
RUN chown postgres /var/lib/postgresql/postgres.server.key && \
    chmod 600 /var/lib/postgresql/postgres.server.key
