CREATE TABLE IF NOT EXISTS players(username text PRIMARY KEY, score INTEGER);
CREATE TABLE IF NOT EXISTS session_words(
    id SERIAL,
    word text UNIQUE,
    username text UNIQUE
);