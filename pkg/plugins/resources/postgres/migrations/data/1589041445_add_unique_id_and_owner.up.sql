ALTER TABLE resources
    DROP CONSTRAINT resources_pkey,
    ADD CONSTRAINT resources_pkey PRIMARY KEY (name, mesh, type);
ALTER TABLE resources
    DROP COLUMN namespace;
ALTER TABLE resources
    ADD COLUMN owner_name varchar(100);
ALTER TABLE resources
    ADD COLUMN owner_mesh varchar(100);
ALTER TABLE resources
    ADD COLUMN owner_type varchar(100);
ALTER TABLE resources
    ADD CONSTRAINT owner_fk FOREIGN KEY (owner_name, owner_mesh, owner_type) REFERENCES resources (name, mesh, type) ON DELETE CASCADE;

DELETE
FROM resources
WHERE type = 'DataplaneInsight';

-- set owner for all resources except Mesh
UPDATE resources r1
SET owner_name = r2.name,
    owner_mesh = r2.mesh,
    owner_type = 'Mesh'
FROM resources r2
WHERE r2.name = r1.mesh
  AND r2.type = 'Mesh'
  AND r1.type != 'Mesh';

--- delete dangling resources (ones without owner)
DELETE
FROM resources
WHERE owner_name IS NULL
  AND owner_mesh IS NULL
  AND owner_type IS NULL
  AND type != 'Mesh';