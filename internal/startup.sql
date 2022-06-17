PRAGMA journal_mode = 'wal';

CREATE TABLE IF NOT EXISTS diary(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL,
    last_modified_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS IX_diary_created_at ON diary(created_at DESC);

CREATE TABLE IF NOT EXISTS diary_content(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    diary_id INTEGER NOT NULL,
    content TEXT NOT NULL COLLATE BINARY,
    comment TEXT NOT NULL COLLATE BINARY,
    created_at INTEGER NOT NULL,
    FOREIGN KEY(diary_id) REFERENCES diary(id)
);

CREATE INDEX IF NOT EXISTS IX_diary_content_created_at ON diary_content(created_at DESC);