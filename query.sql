-- USERS --

-- name: ListUsers :many
SELECT * FROM users;

-- name: GetUserById :one
SELECT * FROM users WHERE Id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE Email = $1;

-- name: InsertUser :one
INSERT INTO users (Name, Email) VALUES ($1, $2) RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE Id = $1;

-- POSTS --

-- name: ListPosts :many
SELECT * FROM posts;

-- name: GetPostById :one
SELECT * FROM posts WHERE Id = $1;

-- name: ListPostsByCreator :many
SELECT * FROM posts WHERE CreatorId = $1;

-- name: ListApprovedPosts :many
SELECT * FROM posts WHERE ApprovedBy IS NOT NULL;

-- name: ListNotApprovedPosts :many
SELECT * FROM posts WHERE ApprovedBy IS NULL;

-- name: CreatePost :one
INSERT INTO posts (Title, Body, CreatorId, ApprovedBy) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE Id = $1;

-- TAGS --

-- name: ListTags :many
SELECT * FROM tags;

-- name: ListTagsByPost :many
SELECT * FROM tags WHERE Id = $1;

-- name: ListPostsByTag :many
SELECT posts.* FROM posts INNER JOIN tags ON posts.Id = tags.PostId WHERE tags.Id = $1;

-- name: CreateTag :one
INSERT INTO tags (PostId, Tag) VALUES ($1, $2) RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tags WHERE Id = $1;

