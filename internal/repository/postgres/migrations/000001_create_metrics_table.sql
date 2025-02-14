-- +goose Up
CREATE TABLE IF NOT EXISTS metric
(
    id   VARCHAR(30) PRIMARY KEY,
    type VARCHAR(20),
    value double precision,
    delta bigint
);

-- +goose Down
DROP TABLE IF EXISTS metric;