CREATE TABLE parameters (
    name TEXT PRIMARY KEY,
    value TEXT
);

CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    type TEXT NOT NULL,
    parameters TEXT
);

CREATE TABLE groups (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE roles (
    id INTEGER PRIMARY KEY,
    gid INTEGER,
    user INTEGER,
    FOREIGN KEY (gid) REFERENCES groups(id),
    FOREIGN KEY (user) REFERENCES users(id)
);

CREATE TABLE acl_users (
    id INTEGER PRIMARY KEY,
    view INTEGER,
    user INTEGER,
    FOREIGN KEY (view) REFERENCES views(id),
    FOREIGN KEY (user) REFERENCES users(id)
);

CREATE TABLE acl_groups (
    id INTEGER PRIMARY KEY,
    view INTEGER,
    gid INTEGER,
    FOREIGN KEY (view) REFERENCES views(id),
    FOREIGN KEY (gid) REFERENCES groups(id)
);

CREATE TABLE sources (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    parameters TEXT,
    json TEXT
);

CREATE TABLE source_require (
    id INTEGER PRIMARY KEY,
    source INTEGER,
    require INTEGER,
    FOREIGN KEY (source) REFERENCES sources(id),
    FOREIGN KEY (require) REFERENCES sources(id)
);

CREATE TABLE items (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    parameters TEXT,
    template TEXT
);

CREATE TABLE item_sources (
    id INTEGER PRIMARY KEY,
    item INTEGER,
    source INTEGER,
    FOREIGN KEY (item) REFERENCES items(id),
    FOREIGN KEY (source) REFERENCES sources(id)
);

CREATE TABLE views (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    parameters TEXT
);

CREATE TABLE view_items (
    id INTEGER PRIMARY KEY,
    view INTEGER,
    item INTEGER,
    parameters TEXT,
    FOREIGN KEY (view) REFERENCES views(id),
    FOREIGN KEY (item) REFERENCES items(id)
);