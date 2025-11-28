-- +goose Up
CREATE TABLE IF NOT EXISTS chartstatscn (
    song_id INTEGER NOT NULL,
    diff TEXT NOT NULL,
    count INTEGER NOT NULL,
    std_dev FLOAT NOT NULL,
    PRIMARY KEY (song_id, diff)
);

-- +goose Down
DROP TABLE chartstatscn;
