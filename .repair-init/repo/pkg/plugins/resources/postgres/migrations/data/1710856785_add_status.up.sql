ALTER TABLE resources ADD COLUMN status JSONB NOT NULL DEFAULT '{}'::jsonb;
