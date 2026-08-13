-- +goose Up
-- +goose StatementBegin
ALTER TABLE documents ADD COLUMN approved_at TIMESTAMP WITH TIME ZONE;

UPDATE documents
SET approved_at = created_at
WHERE status = 'approved' AND approved_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE documents DROP COLUMN approved_at;
-- +goose StatementEnd
