-- +goose Up
CREATE TABLE IF NOT EXISTS songdatacn (
    id INTEGER NOT NULL,
    diff TEXT NOT NULL,
    title TEXT NOT NULL,
    is_dx BOOLEAN NOT NULL,
    PRIMARY KEY (id, diff)
);

-- +goose Down
DROP TABLE songdatacn;
