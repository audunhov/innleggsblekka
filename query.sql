-- USERS --

-- name: ListUsers :many
SELECT * FROM users;

-- name: GetUserById :one
SELECT * FROM users WHERE Id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE Email = ?;

-- name: InsertUser :one
INSERT INTO users (Name, Email) VALUES (?, ?) RETURNING *;

-- name: UpdateUser :one
UPDATE users SET Name=?, Email=? RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE Id = ?;

-- POSTS --

-- name: ListPosts :many
SELECT * FROM posts;

-- name: SearchPosts :many
SELECT *, highlight(posts_fts, 2, '<b>', '</b>') as highlight FROM posts_fts WHERE posts_fts MATCH ? ORDER BY rank;

-- name: GetPostById :one
SELECT * FROM posts WHERE Id = ?;

-- name: ListPostsByCreator :many
SELECT * FROM posts WHERE CreatorId = ?;

-- name: ListApprovedPosts :many
SELECT * FROM posts WHERE ApprovedBy IS NOT NULL;

-- name: ListNotApprovedPosts :many
SELECT * FROM posts WHERE ApprovedBy IS NULL;

-- name: CreatePost :one
INSERT INTO posts (Title, Body, CreatorId, ApprovedBy) VALUES (?, ?, ?, ?) RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE Id = ?;

-- name: ApprovePost :one
UPDATE posts SET ApprovedBy=?, ApprovedAt=current_timestamp WHERE Id = ? RETURNING *;

-- name: UpdatePost :one
UPDATE posts SET Title=?, Body=? WHERE Id=? RETURNING *;

-- name: RemoveApprovalFromPost :one
UPDATE posts SET ApprovedBy=NULL, ApprovedAt=NULL WHERE Id = ? RETURNING *;

-- TAGS --

-- name: ListTags :many
SELECT * FROM tags;

-- name: ListTagsWithPostCount :many
SELECT sqlc.embed(tags), count(post_tags.PostId) as count FROM tags INNER JOIN post_tags on tags.Id = post_tags.TagId GROUP BY tags.Id ORDER BY count DESC;

-- name: GetTagById :one
SELECT * FROM tags WHERE Id = ?;

-- name: ListPostsByTag :many
SELECT posts.* FROM posts INNER JOIN post_tags ON posts.Id = post_tags.PostId WHERE post_tags.TagId = ?;

-- name: ListTagsByPost :many
SELECT tags.* FROM tags INNER JOIN post_tags ON tags.Id = post_tags.TagId WHERE post_tags.PostId = ?;

-- name: UpdateTag :one
UPDATE tags SET Tag=? RETURNING *;

-- name: CreateTag :one
INSERT INTO tags (Tag) VALUES (?) RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tags WHERE Id = ?;

-- name: AddTagToPost :one
INSERT INTO post_tags (TagId, PostId) VALUES (?, ?) RETURNING *;

-- LOG_ENTRIES --

-- name: ListLogEntries :many
SELECT * FROM log_entries;

-- name: GetLogEntryById :one
SELECT * FROM log_entries WHERE Id = ?;

-- name: ListLogEntriesByUser :many
SELECT * FROM log_entries WHERE User = ?;

-- name: ListLogEntriesByTable :many
SELECT * FROM log_entries WHERE TableName = ?;

-- name: ListLogEntriesByUserAndTable :many
SELECT * FROM log_entries WHERE User = ? AND TableName = ?;

-- name: CreateLogEntry :one
INSERT INTO log_entries (TableName, User, OldValue, NewValue) VALUES (?, ?, ?, ?) RETURNING *;

-- name: DeleteLogEntry :exec
DELETE FROM log_entries WHERE Id = ?;

-- name: ListFavouritePostsForUser :many
SELECT posts.* from posts INNER JOIN favourite_posts ON posts.Id = favourite_posts.PostId;

-- name: ListMostFavouritedPosts :many
SELECT posts.Title, count(fp.PostId) as count from posts INNER JOIN favourite_posts as fp ON posts.Id = fp.PostId GROUP BY fp.PostId ORDER BY count DESC; 


-- name: AddFavourite :one
INSERT INTO favourite_posts (PostId, UserId) VALUES (?, ?) RETURNING *;

-- name: RemoveFavourite :exec
DELETE FROM favourite_posts WHERE PostId = ? AND UserId = ?;
