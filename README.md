# Unloaded service

``` sql 

CREATE TABLE anime (
    id BIGSERIAL PRIMARY KEY,
    mal_id INTEGER UNIQUE NOT NULL,
    url TEXT NOT NULL,
    images JSONB,
    title TEXT NOT NULL,
    title_english TEXT,
    type TEXT,
    episodes INTEGER,
    status TEXT,
    rating TEXT,
    score DOUBLE PRECISION,
    synopsis TEXT,
    year INTEGER
);

CREATE TABLE genres (
    id BIGSERIAL PRIMARY KEY,
    mal_id INTEGER UNIQUE NOT NULL,
    type TEXT,
    name TEXT NOT NULL,
    url TEXT
);

CREATE TABLE anime_genres (
    anime_id BIGINT REFERENCES anime(id) ON DELETE CASCADE,
    genre_id BIGINT REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (anime_id, genre_id)
);

```