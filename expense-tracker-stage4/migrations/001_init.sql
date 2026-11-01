CREATE TABLE IF NOT EXISTS categories (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT    NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS expenses (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    amount      REAL    NOT NULL,
    description TEXT    NOT NULL,
    date        TEXT    NOT NULL,
    category_id INTEGER NOT NULL REFERENCES categories(id)
);

INSERT OR IGNORE INTO categories (id, name) VALUES
    (1, 'Еда'),
    (2, 'Транспорт'),
    (3, 'Жильё'),
    (4, 'Развлечения'),
    (5, 'Прочее');