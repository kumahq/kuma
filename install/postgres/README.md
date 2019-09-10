## Postgres installation

You can install Kuma on Postgres DB with following command:
```bash
kumactl install postgres-schema | PGPASSWORD=mysecretpassword psql -h localhost -U postgres
```

### Schema

| Column    |      Type     |  Description                                    |
|-----------|---------------|-------------------------------------------------|
| NAME      |  varchar(100) | Name of the resource                            |
| NAMESPACE |  varchar(100) | Namespace for which the resource belongs to     |
| MESH      |  varchar(100) | Mesh for which the resource belongs to          |
| TYPE      |  varchar(100) | Type of resource                                |
| VERSION   |  integer      | Version for optimistic locking                  |
| SPEC      |  text         | Specification (content) of the resource in JSON |