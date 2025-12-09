-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT * FROM posts
INNER JOIN feed
ON posts.feed_id = feed.id
WHERE $1 = feed.user_id
ORDER BY published_at DESC
LIMIT $2;

-- name: GetPostFromUrl :one
SELECT * FROM posts
WHERE $1 = posts.url;