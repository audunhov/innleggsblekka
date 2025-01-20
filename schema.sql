CREATE TABLE users (
Id INTEGER PRIMARY KEY,
	Name TEXT NOT NULL,
	Email TEXT NOT NULL UNIQUE,
	CreatedAt TIMESTAMP DEFAULT current_timestamp,
	IsAdmin BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE posts (
Id INTEGER PRIMARY KEY,
	Title TEXT NOT NULL,
	Body TEXT NOT NULL,
	CreatorId INTEGER NOT NULL REFERENCES users(Id),
	CreatedAt TIMESTAMP DEFAULT current_timestamp,
	ApprovedBy INTEGER REFERENCES users(Id),
	ApprovedAt TIMESTAMP
);

CREATE VIRTUAL TABLE posts_fts USING fts5 (Id, Title, Body);
CREATE TRIGGER insert_posts_fts AFTER INSERT ON posts BEGIN INSERT INTO posts_fts(Id, Title, Body) VALUES (New.Id, New.Title, New.Body); end;
CREATE TRIGGER update_posts_fts AFTER UPDATE ON posts BEGIN UPDATE posts_fts SET Tile = New.Title, Body = New.Body WHERE Id=New.Id; end;
CREATE TRIGGER delete_posts_fts AFTER DELETE ON posts BEGIN DELETE FROM posts_fts WHERE Id=Old.Id; end;


CREATE TABLE tags (
	Id INTEGER PRIMARY KEY,
	CreatedAt TIMESTAMP DEFAULT current_timestamp,
	Tag TEXT NOT NULL UNIQUE
);

CREATE TABLE post_tags (
	TagId INTEGER NOT NULL REFERENCES tags(Id),
PostId INTEGER NOT NULL REFERENCES posts(Id),
	CreatedAt TIMESTAMP DEFAULT current_timestamp,
	PRIMARY KEY(TagId, PostId)
	);


CREATE TABLE log_entries (
	Id INTEGER PRIMARY KEY,
	TableName TEXT NOT NULL,
	User INTEGER NOT NULL REFERENCES users(Id),
	OldValue BLOB,
	NewValue BLOB,
	CreatedAt TIMESTAMP DEFAULT current_timestamp
);

CREATE TABLE favourite_posts (
	PostId INTEGER NOT NULL REFERENCES posts(Id),
	UserId INTEGER NOT NULL REFERENCES users(Id),
	CreatedAt TIMESTAMP DEFAULT current_timestamp,
	PRIMARY KEY(PostId, UserId)
);
