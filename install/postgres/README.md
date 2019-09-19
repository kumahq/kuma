## Postgres installation

Run following command for installation examples:
```bash
kumactl install database-schema --help
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