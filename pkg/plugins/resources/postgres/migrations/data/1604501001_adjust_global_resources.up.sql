-- Global resources (Mesh and Zone for now) needs to have an empty value in mesh column. This migration updates those resources
ALTER TABLE resources DROP CONSTRAINT owner_fk; -- drop constraint so we can update resources;
UPDATE resources SET mesh = '' WHERE type = 'Mesh' OR type = 'Zone';
UPDATE resources SET owner_mesh = '' WHERE owner_type = 'Mesh' OR owner_type = 'Zone';
ALTER TABLE resources ADD CONSTRAINT owner_fk FOREIGN KEY (owner_name, owner_mesh, owner_type) REFERENCES resources (name, mesh, type) ON DELETE CASCADE; -- apply the constraint again;

