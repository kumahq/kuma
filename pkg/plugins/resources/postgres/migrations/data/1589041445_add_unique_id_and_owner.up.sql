ALTER TABLE resources
    DROP CONSTRAINT resources_pkey,
    ADD CONSTRAINT resources_pkey PRIMARY KEY (name, mesh, type);
ALTER TABLE resources DROP COLUMN namespace;
ALTER TABLE resources ADD COLUMN owner_name varchar(100);
ALTER TABLE resources ADD COLUMN owner_mesh varchar(100);
ALTER TABLE resources ADD COLUMN owner_type varchar(100);
ALTER TABLE resources ADD CONSTRAINT owner_fk FOREIGN KEY (owner_name,owner_mesh,owner_type) REFERENCES resources(name,mesh,type) ON DELETE CASCADE;