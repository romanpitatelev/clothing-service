-- +migrate Up
CREATE TABLE users
(
    id          UUID PRIMARY KEY,
    first_name  VARCHAR   NOT NULL,
    last_name   VARCHAR   NOT NULL,
    nick_name   VARCHAR   NOT NULL,
    gender      VARCHAR   NOT NULL,
    age         INTEGER   NOT NULL CHECK (age > 12),
    email       VARCHAR   NOT NULL UNIQUE,
    phone       VARCHAR   NOT NULL UNIQUE,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_verified BOOLEAN   NOT NULL DEFAULT FALSE,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP WITH TIME ZONE
);

CREATE TABLE registrations
(
    phone      VARCHAR,
    secret     VARCHAR NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE users;
DROP TABLE registrations;

