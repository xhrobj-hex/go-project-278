-- name: ListLinks :many
SELECT
    id,
    original_url,
    short_name,
    created_at
FROM links
ORDER BY id DESC;
