-- name: CreateSong :exec
INSERT INTO songdata (
    id, title, artist, genre, img, release, version, is_dx, diff, level, const, is_utage, is_buddy
)
VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT (id, diff, is_dx, is_utage) DO NOTHING;


-- name: ReturnAllJackets :many
SELECT img FROM songdata;

-- name: CreateSongCN :exec
INSERT INTO songdatacn (
    id, diff, title, is_dx
)
VALUES (
    ?, ?, ?, ?
)
ON CONFLICT (id, diff) DO NOTHING;

-- name: CreateChartStatCN :exec
INSERT INTO chartstatscn (
    song_id, diff, count, std_dev
)
VALUES (
    ?, ?, ?, ?
)
ON CONFLICT (song_id, diff) DO NOTHING;
