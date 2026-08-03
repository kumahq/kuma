CREATE TABLE locks (
    name CHARACTER VARYING(255) PRIMARY KEY,
    record_version_number BIGINT,
    data BYTEA,
    owner CHARACTER VARYING(255)
);

CREATE SEQUENCE locks_rvn OWNED BY locks.record_version_number;
