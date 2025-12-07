-- name: MarkFeedFetched :one
UPDATE feed
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE $1 = id
RETURNING *;

-- name: GetNextFeedToFetch :one
SELECT * FROM feed
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;