-- +migrate Up
CREATE TABLE users
(
    id             UUID PRIMARY KEY,
    first_name     VARCHAR,
    last_name      VARCHAR,
    nick_name      VARCHAR                  NOT NULL,
    gender         VARCHAR,
    birth_date     DATE,
    email          VARCHAR,
    email_verified BOOLEAN                  NOT NULL DEFAULT FALSE,
    phone          VARCHAR                  NOT NULL,
    phone_verified BOOLEAN                  NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at     TIMESTAMP WITH TIME ZONE,
    otp            VARCHAR,
    otp_created_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX users_unique_phone_and_deleted_at_null_idx
    ON users (phone)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX users_unique_email_and_deleted_at_null_idx
    ON users (email)
    WHERE deleted_at IS NULL;

-- +migrate Down
DROP INDEX users_unique_phone_and_deleted_at_null_idx;
DROP INDEX users_unique_email_and_deleted_at_null_idx;
DROP TABLE users;
