-- name: ListLinks :many
SELECT
    id,
    original_url,
    short_name,
    created_at
FROM links
ORDER BY id DESC;

-- name: CreateLink :one
INSERT INTO links (
    original_url,
    short_name
)
VALUES (
    $1,
    $2
)
RETURNING
    id,
    original_url,
    short_name,
    created_at;

-- name: GetLinkByID :one
SELECT
    id,
    original_url,
    short_name,
    created_at
FROM links
WHERE id = $1;
