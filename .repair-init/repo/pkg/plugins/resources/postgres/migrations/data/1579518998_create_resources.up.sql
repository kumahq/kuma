CREATE TABLE IF NOT EXISTS resources (
    name        varchar(100) NOT NULL,
    namespace   varchar(100) NOT NULL,
    mesh        varchar(100) NOT NULL,
    type        varchar(100) NOT NULL,
    version     integer NOT NULL,
    spec        text,
    PRIMARY KEY (name, namespace, mesh, type)
);