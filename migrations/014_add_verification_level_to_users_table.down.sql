-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

ALTER TABLE users DROP COLUMN IF EXISTS verification_level;