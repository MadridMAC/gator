-- name: CreateFeed :one
INSERT INTO feed (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeeds :many
SELECT name, url, user_id FROM feed;

-- name: GetFeedUser :one
SELECT users.name
FROM feed
INNER JOIN users
ON feed.user_id = users.id
WHERE $1 = users.id;

-- name: GetFeedByUrl :one
SELECT * from feed
WHERE $1 = feed.url;