CREATE TABLE rooms
(
    id      SERIAL PRIMARY KEY,
    site_id SERIAL REFERENCES sites (id) ON DELETE CASCADE,
    name    TEXT
);