-- +goose Up
-- +goose StatementBegin
CREATE TABLE tasks (
    id INTEGER NOT NULL PRIMARY KEY,
    name VARCHAR(255),
    completed BOOLEAN
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tasks;
-- +goose StatementEnd
